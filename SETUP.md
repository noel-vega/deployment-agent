# Setup Guide

Complete installation and configuration guide for Hubble.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Environment Variables](#environment-variables)
- [Registry Setup](#registry-setup)
- [Traefik Setup](#traefik-setup)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Installation

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/noel-vega/hubble
cd hubble

# Copy environment template
cp .env.example .env

# Edit .env with your values
nano .env

# Start Hubble
docker-compose up -d

# View logs
docker-compose logs -f hubble-server
```

### Option 2: Build from Source

```bash
# Prerequisites
# - Go 1.24+
# - Docker installed and running

# Clone and build
git clone https://github.com/noel-vega/hubble
cd hubble
go mod download
go build -o hubble-server .

# Run
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=yourpassword
./hubble-server
```

### Option 3: Development Mode

```bash
# Uses hot reload with Air
make dev

# Or manually
go install github.com/air-verse/air@latest
air
```

## Configuration

### Required Environment Variables

These **must** be set or Hubble will not start:

```bash
# Admin credentials (required)
ADMIN_USERNAME=your-admin-username
ADMIN_PASSWORD=your-secure-password  # min 8 characters

# JWT secrets (required for production)
JWT_ACCESS_SECRET=your-random-secret-min-32-chars
JWT_REFRESH_SECRET=your-different-random-secret-min-32-chars
```

**Generate strong secrets:**

```bash
# Linux/macOS
openssl rand -base64 32

# Or use any password generator
```

### Optional Environment Variables

```bash
# Token duration
ACCESS_TOKEN_DURATION=5m          # How long before re-auth required
REFRESH_TOKEN_DURATION=168h       # How long before login required

# Environment
ENVIRONMENT=development           # Set to 'production' for HTTPS cookies

# Projects directory
PROJECTS_ROOT_PATH=/projects      # Where compose projects are stored

# Platform domain (for HTTPS access)
HUBBLE_DOMAIN=yourdomain.com      # Required for registry and Traefik HTTPS

# External Registry browsing (optional)
REGISTRY_URL=http://localhost:5001
REGISTRY_USERNAME=
REGISTRY_PASSWORD=
```

## Environment Variables

### Authentication

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ADMIN_USERNAME` | ✅ Yes | - | Admin username |
| `ADMIN_PASSWORD` | ✅ Yes | - | Admin password (min 8 chars) |
| `JWT_ACCESS_SECRET` | ⚠️ Recommended | Generated | Access token secret (32+ chars) |
| `JWT_REFRESH_SECRET` | ⚠️ Recommended | Generated | Refresh token secret (32+ chars) |
| `ACCESS_TOKEN_DURATION` | No | `5m` | Access token lifetime |
| `REFRESH_TOKEN_DURATION` | No | `168h` | Refresh token lifetime |
| `ENVIRONMENT` | No | `development` | Set to `production` for HTTPS |

### Projects

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PROJECTS_ROOT_PATH` | No | `/projects` | Root directory for projects |

### Platform

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HUBBLE_DOMAIN` | ⚠️ For HTTPS | - | Base domain for platform services |

### Traefik

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HUBBLE_TRAEFIK_ENABLED` | No | `false` | Enable Traefik auto-provisioning |
| `HUBBLE_TRAEFIK_EMAIL` | ⚠️ If HTTPS | - | Email for Let's Encrypt certs |
| `HUBBLE_TRAEFIK_DASHBOARD` | No | `false` | Enable Traefik dashboard |
| `HUBBLE_TRAEFIK_DASHBOARD_AUTH` | ⚠️ Recommended | - | Dashboard auth (htpasswd format) |

### Hubble Registry (Core Feature)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HUBBLE_REGISTRY_ENABLED` | No | `true` | Enable built-in Docker registry |
| `HUBBLE_REGISTRY_DELETE_ENABLED` | No | `true` | Allow image deletion |
| `HUBBLE_REGISTRY_STORAGE` | No | `/var/lib/hubble/registry` | Image storage path |

### External Registry (Optional)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REGISTRY_URL` | No | - | External registry URL for browsing |
| `REGISTRY_USERNAME` | No | - | External registry username |
| `REGISTRY_PASSWORD` | No | - | External registry password |

## Registry Setup

Hubble includes a built-in Docker Registry for storing your private images. It's enabled by default and integrates seamlessly with Traefik for HTTPS access.

### Enable Registry (Already Default)

The registry is **enabled by default**. To disable it:

```bash
# In .env file
HUBBLE_REGISTRY_ENABLED=false
```

### Configure Domain for HTTPS Access

For HTTPS access to your registry, set your domain:

```bash
# In .env file
HUBBLE_DOMAIN=yourdomain.com
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
```

This makes your registry available at: `https://registry.yourdomain.com`

### Login to Registry

```bash
# Login using your Hubble admin credentials
docker login registry.yourdomain.com
Username: admin
Password: [your-hubble-password]
```

### Push Images

```bash
# 1. Build your image
docker build -t myapp:v1.0 .

# 2. Tag for Hubble registry
docker tag myapp:v1.0 registry.yourdomain.com/myapp:v1.0

# 3. Push
docker push registry.yourdomain.com/myapp:v1.0
```

### Use in Projects

```bash
# Deploy image from your registry
curl -X POST http://localhost:3000/projects/myapp/services \
  -b cookies.txt \
  -d '{
    "name": "web",
    "image": "registry.yourdomain.com/myapp:v1.0",
    "networks": ["hubble"]
  }'
```

### Registry Storage

Images are stored in `/var/lib/hubble/registry` by default. To change:

```bash
# In .env file
HUBBLE_REGISTRY_STORAGE=/custom/path/to/registry
```

### Verify Registry

```bash
# Check registry container
docker ps | grep hubble-registry

# Check registry logs
docker logs hubble-registry

# Test registry API
curl https://registry.yourdomain.com/v2/_catalog \
  -u admin:your-password
```

## Traefik Setup

Traefik provides automatic HTTPS and routing for your projects.

### Enable Traefik

```bash
# In .env file
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
HUBBLE_TRAEFIK_DASHBOARD=true
```

### Generate Dashboard Auth

```bash
# Install htpasswd (if not available)
sudo apt-get install apache2-utils  # Debian/Ubuntu
brew install httpd                   # macOS

# Generate password
htpasswd -nb admin your-password

# Output example: admin:$apr1$xyz123$abc...
# Copy to .env:
HUBBLE_TRAEFIK_DASHBOARD_AUTH=admin:$apr1$xyz123$abc...
```

### Restart Hubble

```bash
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs -f hubble-server
```

You should see:

```
✓ Hubble network 'hubble' already exists (ID: abc123)
Creating Traefik container...
Let's Encrypt configured with email: admin@yourdomain.com
✓ Created and started Traefik (ID: def456)
  HTTP: http://0.0.0.0:80
  HTTPS: https://0.0.0.0:443
  Dashboard available at: http://localhost:8080
```

### Access Dashboard

```bash
# Traefik dashboard is on localhost only (for security)
# SSH tunnel if remote:
ssh -L 8080:localhost:8080 user@your-server

# Then visit:
http://localhost:8080
```

## Production Deployment

### 1. Server Requirements

- **OS**: Linux (Ubuntu 22.04+ recommended)
- **RAM**: 1GB minimum, 2GB+ recommended
- **Storage**: 20GB+ for images and projects
- **Ports**: 80, 443, 3000 (or custom)

### 2. Secure the Server

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt-get install docker-compose-plugin

# Setup firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### 3. Deploy Hubble

```bash
# Clone repository
git clone https://github.com/noel-vega/hubble /opt/hubble
cd /opt/hubble

# Configure production environment
cp .env.example .env
nano .env
```

**Production .env example:**

```bash
# Authentication
ENVIRONMENT=production
ADMIN_USERNAME=admin
ADMIN_PASSWORD=YourVerySecurePassword123!
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# Traefik
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
HUBBLE_TRAEFIK_DASHBOARD=true
HUBBLE_TRAEFIK_DASHBOARD_AUTH=$(htpasswd -nb admin yourpassword)

# Projects
PROJECTS_ROOT_PATH=/opt/hubble/projects
```

### 4. Start Services

```bash
# Create projects directory
sudo mkdir -p /opt/hubble/projects
sudo chown $USER:$USER /opt/hubble/projects

# Start Hubble
docker-compose up -d

# Enable auto-restart on boot
docker-compose restart
```

### 5. Verify Deployment

```bash
# Check container status
docker ps

# Check Hubble logs
docker-compose logs -f hubble-server

# Check Traefik
docker logs hubble-traefik

# Test API
curl http://localhost:3000/
```

### 6. DNS Configuration

Point your domain to your server's IP:

```
Type    Name    Value           TTL
A       @       YOUR.SERVER.IP  300
A       *       YOUR.SERVER.IP  300
```

This enables:
- `yourdomain.com` → Hubble API
- `*.yourdomain.com` → Your projects

## Troubleshooting

### Hubble Won't Start

**Error**: "Failed to initialize users: missing credentials"

```bash
# Solution: Set required environment variables
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=yourpassword
```

**Error**: "Failed to initialize docker service"

```bash
# Solution: Ensure Docker is running
sudo systemctl start docker

# Verify Docker socket
ls -l /var/run/docker.sock
```

### Traefik Not Creating

**Error**: "Failed to create Traefik container: port already in use"

```bash
# Check what's using port 80/443
sudo lsof -i :80
sudo lsof -i :443

# Stop conflicting service
sudo systemctl stop nginx  # or apache2
```

### Authentication Not Working

**Error**: "Unauthorized - No access token"

```bash
# Ensure cookies are being sent
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpass"}' \
  -c cookies.txt -v

# Check cookie file was created
cat cookies.txt

# Use cookie in subsequent requests
curl http://localhost:3000/projects -b cookies.txt
```

### HTTPS Not Working

**Error**: Traefik not issuing certificates

```bash
# Check Traefik logs
docker logs hubble-traefik

# Verify email is set
echo $HUBBLE_TRAEFIK_EMAIL

# Ensure ports 80/443 are accessible from internet
curl -I http://yourdomain.com
curl -I https://yourdomain.com

# Check Let's Encrypt rate limits
# https://letsencrypt.org/docs/rate-limits/
```

### Projects Not Starting

**Error**: "network hubble not found"

```bash
# Recreate Hubble network
docker network rm hubble
docker network create hubble --label com.hubble.managed=true

# Or restart Hubble to auto-create
docker-compose restart hubble-server
```

### Docker Socket Permission Denied

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or:
newgrp docker

# Verify
docker ps
```

## Upgrading

```bash
cd /opt/hubble

# Pull latest changes
git pull

# Rebuild
docker-compose build

# Restart
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs -f
```

## Backup

### Backup Projects

```bash
# Backup all project data
tar -czf hubble-projects-backup.tar.gz /opt/hubble/projects

# Backup Traefik data (certificates)
tar -czf hubble-traefik-backup.tar.gz /var/lib/hubble/traefik
```

### Backup Environment

```bash
# Backup .env file (contains secrets!)
cp .env .env.backup

# Store securely (DO NOT commit to git)
```

### Restore

```bash
# Restore projects
tar -xzf hubble-projects-backup.tar.gz -C /

# Restore Traefik data
tar -xzf hubble-traefik-backup.tar.gz -C /

# Restart services
docker-compose restart
```

## Uninstall

```bash
# Stop all services
docker-compose down

# Remove containers
docker-compose rm -f

# Remove images
docker rmi hubble_hubble-server

# Remove volumes (if any)
docker volume prune

# Remove Traefik
docker rm -f hubble-traefik

# Remove network
docker network rm hubble

# Remove files
rm -rf /opt/hubble
rm -rf /var/lib/hubble
```

## Next Steps

- **[API.md](API.md)** - Learn the API
- **[TRAEFIK.md](TRAEFIK.md)** - Configure Traefik routing
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Contribute to Hubble
