# Hubble

**The complete self-hosted Docker platform with declarative infrastructure.**

Hubble is a production-ready platform for managing Docker containers and Compose projects. Deploy apps with automatic HTTPS, store images in your own registry, and manage everything through a simple REST API.

## Features

- **Declarative Infrastructure** - All services defined in docker-compose.yml (v2.0)
- **Built-in Docker Registry** - Private image storage with automatic HTTPS
- **Docker Volume Management** - Persistent storage with zero permission issues
- **Project Management API** - Create and manage Docker Compose projects via REST API
- **Traefik Integration** - Automatic HTTPS and routing with Let's Encrypt
- **Secure Authentication** - JWT-based auth with refresh token rotation
- **Container Monitoring** - List, start, stop, and inspect containers
- **React Web UI** - Modern frontend for managing your infrastructure

## Quick Start

**New to Hubble?** Follow our [**step-by-step walkthrough**](WALKTHROUGH.md) for a complete guided setup (30 minutes).

### 1. Clone and Configure

```bash
git clone https://github.com/noel-vega/hubble
cd hubble
cp .env.example .env
```

Edit `.env` and set required values:

```bash
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# For production with HTTPS
HUBBLE_DOMAIN=yourdomain.com
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
```

### 2. Start Hubble

```bash
docker compose up -d
```

This starts all services:
- ✅ **hubble-traefik** - Reverse proxy with automatic HTTPS
- ✅ **hubble-registry** - Private Docker registry
- ✅ **hubble-server** - REST API backend
- ✅ **hubble-web** - React frontend UI
- ✅ **hubble network** - Shared container network

### 3. Login

```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-secure-password"}' \
  -c cookies.txt
```

### 4. Create a Project

```bash
curl -X POST http://localhost:3000/projects \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"my-blog"}'
```

### 5. Add a Service

```bash
curl -X POST http://localhost:3000/projects/my-blog/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "web",
    "image": "nginx:alpine",
    "ports": ["80:80"],
    "restart": "unless-stopped"
  }'
```

## What is Hubble?

Hubble is a **complete platform** for self-hosted Docker infrastructure. It provides:

### Declarative Infrastructure (v2.0)
- All services defined in `docker-compose.yml`
- Shared `hubble` Docker network
- Docker Registry for private image storage
- Traefik reverse proxy with automatic HTTPS
- All managed services labeled with `com.hubble.managed=true`

### Built-in Registry
- Private Docker registry at `registry.yourdomain.com`
- Automatic HTTPS via Let's Encrypt (when Traefik enabled)
- Uses your Hubble admin credentials
- Store unlimited images on your infrastructure

### Project Management
- Create Docker Compose projects via API
- Add/update/delete services and networks
- Start/stop services individually or as a group
- All projects stored in `/projects` directory

### Zero-Config Networking
- All projects automatically connect to the `hubble` network
- Registry and Traefik integrated seamlessly
- No manual network or infrastructure commands needed

### Complete Workflow
```bash
# 1. Build your app
docker build -t myapp .

# 2. Push to Hubble registry
docker tag myapp registry.yourdomain.com/myapp
docker push registry.yourdomain.com/myapp

# 3. Deploy via API
curl -X POST /projects/myapp/services \
  -d '{"name":"web","image":"registry.yourdomain.com/myapp"}'

# 4. Access at myapp.yourdomain.com (via Traefik)
```

## Architecture

### v2.0 Service Architecture
```
┌─────────────────────────────────────────────────────────┐
│ docker-compose.yml (ALL infrastructure)                 │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  hubble-traefik     → HTTPS reverse proxy                │
│  hubble-registry    → Docker image storage               │
│  hubble-server      → API backend (Go)                   │
│  hubble-web         → React frontend                     │
│  registry-init      → One-time auth setup                │
│                                                           │
│  hubble network     → Shared container network           │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### Request Flow
```
Internet → Traefik → Hubble Network
                          ↓
           ┌──────────────┼──────────────┐
           ↓              ↓              ↓
    hubble-web     hubble-server   hubble-registry
    (Frontend)        (API)         (Images)
```

### Persistent Storage

Hubble uses Docker-managed volumes for all persistent data:

- `hubble-registry-data` - Docker images and layers
- `hubble-registry-auth` - Registry authentication
- `hubble-traefik-data` - Traefik configuration and SSL certificates
- `hubble-projects` - Docker Compose project files

**Benefits:**
- ✅ Zero permission issues - Docker handles all access
- ✅ No manual setup required
- ✅ Easy backup and restore
- ✅ Portable across systems

## Use Cases

- **Complete Self-Hosted Platform** - Everything you need: API, registry, routing, HTTPS, web UI
- **Private Image Storage** - No Docker Hub rate limits, full control of your images
- **Dev/Staging Environment** - Quick project setup with declarative infrastructure
- **Learning Docker** - REST API + Web UI for Docker Compose operations
- **Automated Deployments** - API-driven infrastructure with built-in registry

## Documentation

- **[WALKTHROUGH.md](WALKTHROUGH.md)** - 🎯 **START HERE!** Step-by-step guide for new users
- **[SETUP.md](SETUP.md)** - Installation, configuration, and deployment
- **[API.md](API.md)** - Complete API reference
- **[TRAEFIK.md](TRAEFIK.md)** - Traefik integration and examples
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development workflow and testing

### Migration Guides
- **[PHASE3_NOTES.md](PHASE3_NOTES.md)** - v2.0 migration (declarative infrastructure)
- **[PHASE2_NOTES.md](PHASE2_NOTES.md)** - v1.1 deprecation warnings
- **[PHASE1_NOTES.md](PHASE1_NOTES.md)** - v1.0 docker-compose integration

## Requirements

- Docker 20.10+
- Docker Compose v2+
- Go 1.24+ (for local development)
- Linux host (recommended) or macOS

## Quick Commands

```bash
# Development
make dev                # Start with hot reload
make test-auth          # Test authentication flow

# Building
make build              # Build binary
make docker-build       # Build Docker image

# Deployment
docker-compose up -d    # Start with Docker Compose
docker-compose logs -f  # View logs
docker-compose down     # Stop everything
```

## Security

- ✅ JWT authentication with httpOnly cookies
- ✅ Bcrypt password hashing
- ✅ Token rotation on refresh
- ✅ HTTPS enforcement in production
- ✅ Minimum 8-character passwords
- ✅ No default credentials

## Environment Variables

Key environment variables (v2.0):

```bash
# Authentication (required)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-password
JWT_ACCESS_SECRET=random-secret-32-chars
JWT_REFRESH_SECRET=different-random-secret

# Platform Domain (for HTTPS access)
HUBBLE_DOMAIN=yourdomain.com

# Traefik (for Let's Encrypt HTTPS)
HUBBLE_TRAEFIK_EMAIL=admin@example.com
HUBBLE_TRAEFIK_DASHBOARD=false

# Registry
HUBBLE_REGISTRY_DELETE_ENABLED=true

# Projects
PROJECTS_ROOT_PATH=/projects
```

**Note:** In v2.0, all infrastructure is defined in docker-compose.yml. Environment variables are used for configuration only, not for enabling/disabling services.

See [SETUP.md](SETUP.md) for complete configuration details.

## License

MIT

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development setup and guidelines.
