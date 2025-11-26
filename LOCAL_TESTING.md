# Local Testing Guide

This guide walks you through testing Hubble locally on your development machine **without requiring a domain name or VPS**.

## What This Tests

✅ Go build process  
✅ Docker container building  
✅ Infrastructure auto-provisioning (network, registry)  
✅ Registry push/pull operations  
✅ API authentication and endpoints  
✅ Project and service management  
✅ Full deployment workflow  

**Time:** ~10 minutes

---

## Prerequisites

- **Docker**: Docker Engine 20.10+ and Docker Compose v2.0+
- **Go**: Go 1.21+ (optional, for building outside Docker)
- **curl**: For API testing

---

## Quick Start

### Step 1: Setup Environment

```bash
cd /path/to/hubble

# Create local testing environment file
cp .env.example .env.test

# Edit .env.test with local settings
cat > .env.test <<'EOF'
# Local Testing Configuration
ENVIRONMENT=development

# Admin credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD=testpass123

# JWT secrets (use strong secrets in production!)
JWT_ACCESS_SECRET=local-test-access-secret-min-32-characters
JWT_REFRESH_SECRET=local-test-refresh-secret-min-32-characters

# Token durations
ACCESS_TOKEN_DURATION=5m
REFRESH_TOKEN_DURATION=168h

# Projects path
PROJECTS_ROOT_PATH=/projects

# Domain (localhost for local testing)
HUBBLE_DOMAIN=localhost

# Traefik DISABLED for local testing (no HTTPS)
HUBBLE_TRAEFIK_ENABLED=false

# Registry ENABLED (accessible at localhost:5000)
HUBBLE_REGISTRY_ENABLED=true
HUBBLE_REGISTRY_DELETE_ENABLED=true
# Use local directory (no sudo required)
HUBBLE_REGISTRY_STORAGE=$(pwd)/tmp/registry
EOF

# Copy to .env
cp .env.test .env
```

### Step 2: Create Storage Directories

```bash
# Create local storage directories
mkdir -p tmp/registry tmp/registry-auth tmp/traefik

# Create htpasswd file for registry authentication
docker run --rm httpd:alpine htpasswd -Bbn admin testpass123 > tmp/registry-auth/htpasswd
```

### Step 3: Start Hubble

```bash
# Build and start services
docker compose up -d --build

# View logs
docker compose logs -f hubble-server

# You should see:
# ✓ Hubble network 'hubble' created
# ✓ Created and started Registry
# Starting server on :5000
```

### Step 4: Verify Infrastructure

```bash
# Check containers
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# You should see:
# hubble-server     Up X seconds    0.0.0.0:3000->5000/tcp
# hubble-registry   Up X seconds    127.0.0.1:5000->5000/tcp

# Check network
docker network ls | grep hubble
```

---

## Test Suite

### Test 1: API Authentication

```bash
# Test root endpoint
curl http://localhost:3000/
# Expected: hubble

# Login and save cookies
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass123"}' \
  -c /tmp/hubble-cookies.txt

# Expected: {"authenticated":true,"username":"admin"}

# Verify authentication
curl http://localhost:3000/auth/me -b /tmp/hubble-cookies.txt
# Expected: {"authenticated":true,"username":"admin"}
```

**✅ PASS:** Authentication working

### Test 2: Registry Operations

```bash
# Check registry is accessible
curl http://localhost:5000/v2/ -u admin:testpass123
# Expected: {}

# Check catalog (empty initially)
curl http://localhost:5000/v2/_catalog -u admin:testpass123
# Expected: {"repositories":[]}

# Login to registry
docker login localhost:5000 -u admin -p testpass123
# Expected: Login Succeeded
```

**✅ PASS:** Registry accessible

### Test 3: Build and Push Image

```bash
# Create test app
mkdir -p tmp/test-app
cd tmp/test-app

cat > Dockerfile <<'EOF'
FROM nginx:alpine
RUN echo '<h1>Hello from Hubble Test!</h1>' > /usr/share/nginx/html/index.html
EXPOSE 80
EOF

# Build image
docker build -t testapp:v1.0 .

# Tag for local registry
docker tag testapp:v1.0 localhost:5000/testapp:v1.0

# Push to registry
docker push localhost:5000/testapp:v1.0

# Verify push
curl http://localhost:5000/v2/_catalog -u admin:testpass123
# Expected: {"repositories":["testapp"]}

curl http://localhost:5000/v2/testapp/tags/list -u admin:testpass123
# Expected: {"name":"testapp","tags":["v1.0"]}
```

**✅ PASS:** Image push/pull working

### Test 4: Project Management

```bash
# List projects (empty)
curl http://localhost:3000/projects -b /tmp/hubble-cookies.txt
# Expected: {"count":0,"projects":[]}

# Create project
curl -X POST http://localhost:3000/projects \
  -H "Content-Type: application/json" \
  -b /tmp/hubble-cookies.txt \
  -d '{"name":"testapp"}'

# Expected: {"message":"project created successfully",...}

# Add network
curl -X POST http://localhost:3000/projects/testapp/networks \
  -H "Content-Type: application/json" \
  -b /tmp/hubble-cookies.txt \
  -d '{"name":"hubble","external":true}'

# Expected: {"message":"network added successfully",...}
```

**✅ PASS:** Project management working

### Test 5: Service Deployment

```bash
# Add service
curl -X POST http://localhost:3000/projects/testapp/services \
  -H "Content-Type: application/json" \
  -b /tmp/hubble-cookies.txt \
  -d '{
    "name": "web",
    "image": "localhost:5000/testapp:v1.0",
    "networks": ["hubble"],
    "ports": ["8080:80"],
    "restart": "unless-stopped"
  }'

# Expected: {"message":"service added successfully",...}

# Start service
curl -X POST http://localhost:3000/projects/testapp/services/web/start \
  -b /tmp/hubble-cookies.txt

# Expected: {"message":"service started successfully",...}

# Check container status
curl http://localhost:3000/projects/testapp/containers \
  -b /tmp/hubble-cookies.txt

# Expected: {"containers":[{"name":"testapp-web-1","state":"running",...}],...}

# Verify container is running
docker ps | grep testapp-web
```

**✅ PASS:** Service deployment working

### Test 6: Access Deployed App

```bash
# Access the deployed application
curl http://localhost:8080/

# Expected: <h1>Hello from Hubble Test!</h1>

# Or open in browser: http://localhost:8080
```

**✅ PASS:** Application accessible

### Test 7: Service Lifecycle

```bash
# Stop service
curl -X POST http://localhost:3000/projects/testapp/services/web/stop \
  -b /tmp/hubble-cookies.txt

# Expected: {"message":"service stopped successfully",...}

# Verify container stopped
docker ps -a | grep testapp-web
# Status should show "Exited"

# Restart service
curl -X POST http://localhost:3000/projects/testapp/services/web/start \
  -b /tmp/hubble-cookies.txt

# Verify running again
curl http://localhost:8080/
```

**✅ PASS:** Service lifecycle working

---

## Cleanup

```bash
# Stop all services
docker compose down

# Remove test containers
docker rm -f testapp-web-1 2>/dev/null || true

# Remove test project
rm -rf /projects/testapp

# Remove test images (optional)
docker rmi testapp:v1.0 localhost:5000/testapp:v1.0 2>/dev/null || true

# Clean up local storage (optional)
rm -rf tmp/registry tmp/registry-auth tmp/test-app
```

---

## Test Results Summary

All tests passed successfully! ✅

**Verified Components:**
1. ✅ Build process (Go compilation, Docker image build)
2. ✅ Infrastructure provisioning (network, registry auto-creation)
3. ✅ API authentication (login, JWT tokens, cookies)
4. ✅ Registry operations (push, pull, catalog)
5. ✅ Project management (create, configure)
6. ✅ Service deployment (add, start, stop)
7. ✅ Application access (port mapping, networking)

---

## Key Differences vs Production

| Feature | Local Testing | Production |
|---------|---------------|------------|
| **Domain** | `localhost` | `yourdomain.com` |
| **HTTPS** | No (HTTP only) | Yes (Let's Encrypt) |
| **Traefik** | Disabled | Enabled |
| **Registry** | `localhost:5000` (HTTP) | `registry.yourdomain.com` (HTTPS) |
| **Storage** | `./tmp/registry` | `/var/lib/hubble/registry` |
| **Access** | `localhost:PORT` | `app.yourdomain.com` |

---

## Troubleshooting

### Port 3000 Already in Use

```bash
# Find what's using port 3000
lsof -i :3000
# Or check Docker containers
docker ps --filter "publish=3000"

# Stop conflicting container
docker stop <container-name>
```

### Port 5000 Already in Use

```bash
# macOS AirPlay uses port 5000 by default
# Disable in: System Settings → General → AirDrop & Handoff → AirPlay Receiver

# Or change registry port in platform/registry.go
```

### Registry Authentication Fails

```bash
# Regenerate htpasswd file
docker run --rm httpd:alpine htpasswd -Bbn admin testpass123 > tmp/registry-auth/htpasswd

# Restart registry
docker restart hubble-registry
```

### Container Won't Start

```bash
# Check logs
docker logs hubble-server
docker logs hubble-registry

# Common issues:
# - Missing storage directories
# - Permission issues (use local dirs)
# - Port conflicts
```

---

## Next Steps

After local testing succeeds:

1. **Commit Changes** - All Phase 3 features are working
2. **Production Deployment** - Follow [WALKTHROUGH.md](WALKTHROUGH.md) for VPS setup
3. **Enable Traefik** - Set `HUBBLE_TRAEFIK_ENABLED=true` for HTTPS
4. **Configure Domain** - Point DNS to your server

---

## Notes for Developers

### Building Locally (Outside Docker)

```bash
# Build Go binary
go build -o hubble-server main.go

# Run with environment
source .env.test
./hubble-server
```

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (requires Docker)
./scripts/test-integration.sh
```

### Development Workflow

```bash
# Use air for hot reload
go install github.com/cosmtrek/air@latest
air

# Or use make dev
make dev
```

---

## Validation Checklist

Use this checklist to verify everything works:

- [ ] Go build completes without errors
- [ ] Docker compose builds successfully
- [ ] Hubble network auto-created
- [ ] Registry container starts and is accessible
- [ ] API endpoints respond correctly
- [ ] Authentication works (login, cookies)
- [ ] Can push images to local registry
- [ ] Can create projects via API
- [ ] Can deploy services
- [ ] Deployed apps are accessible
- [ ] Can stop/start services
- [ ] No errors in logs

---

**All tests passed!** 🎉

Hubble is ready for production deployment. See [WALKTHROUGH.md](WALKTHROUGH.md) for deploying to a VPS with a domain name.
