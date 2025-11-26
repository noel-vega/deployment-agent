# New User Walkthrough

**Welcome to Hubble!** This guide will walk you through setting up and using Hubble from scratch.

## What You'll Build

By the end of this walkthrough, you'll have:
- ✅ Hubble platform running with full HTTPS
- ✅ Private Docker registry at `registry.yourdomain.com`
- ✅ A deployed application at `myapp.yourdomain.com`
- ✅ Complete CI/CD workflow (build → push → deploy)

**Time:** ~30 minutes

---

## Prerequisites

- **Server**: Linux VPS or dedicated server (1GB+ RAM)
- **Domain**: Domain name pointed to your server
- **Access**: SSH access to your server
- **Docker**: Docker and Docker Compose installed

---

## Part 1: Server Setup (5 minutes)

### 1.1 Connect to Your Server

```bash
ssh user@your-server-ip
```

### 1.2 Install Docker (if not installed)

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose plugin
sudo apt-get install docker-compose-plugin

# Log out and back in for group changes to take effect
exit
ssh user@your-server-ip
```

### 1.3 Verify Docker Installation

```bash
docker --version
docker compose version
```

Expected output:
```
Docker version 24.0+
Docker Compose version v2.20+
```

---

## Part 2: DNS Configuration (2 minutes)

Configure your domain's DNS to point to your server.

### 2.1 Get Your Server IP

```bash
curl ifconfig.me
```

Copy the IP address (e.g., `123.45.67.89`)

### 2.2 Add DNS Records

Go to your domain provider's DNS settings and add:

```
Type    Name    Value           TTL
A       @       123.45.67.89    300
A       *       123.45.67.89    300
```

This enables:
- `yourdomain.com` → Your site
- `registry.yourdomain.com` → Hubble Registry
- `*.yourdomain.com` → All your apps

### 2.3 Verify DNS Propagation

```bash
# Check if DNS is working
dig +short yourdomain.com

# Should return your server IP
```

**Note:** DNS propagation can take 5-60 minutes. Continue while it propagates.

---

## Part 3: Install Hubble (5 minutes)

### 3.1 Clone Hubble

```bash
cd ~
git clone https://github.com/noel-vega/hubble
cd hubble
```

### 3.2 Configure Environment

```bash
# Copy example config
cp .env.example .env

# Edit configuration
nano .env
```

**Minimum required configuration:**

```bash
# Admin credentials (CHANGE THESE!)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=YourSecurePassword123!

# JWT secrets (generate strong random strings)
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# Your domain
HUBBLE_DOMAIN=yourdomain.com

# Enable Traefik for HTTPS
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com

# Registry is enabled by default
HUBBLE_REGISTRY_ENABLED=true
```

**Pro tip:** Run these commands to generate secrets:

```bash
echo "JWT_ACCESS_SECRET=$(openssl rand -base64 32)"
echo "JWT_REFRESH_SECRET=$(openssl rand -base64 32)"
```

Copy the output into your `.env` file.

Save and exit (Ctrl+X, then Y, then Enter)

### 3.3 Start Hubble

```bash
docker-compose up -d
```

### 3.4 View Startup Logs

```bash
docker-compose logs -f hubble-server
```

You should see:
```
✓ Hubble network 'hubble' created (ID: abc123)
Creating Traefik container...
✓ Created and started Traefik (ID: def456)
  HTTP: http://0.0.0.0:80
  HTTPS: https://0.0.0.0:443
  Dashboard: http://localhost:8080
Creating Registry container...
✓ Created and started Registry (ID: ghi789)
  Registry URL: https://registry.yourdomain.com
  Login: docker login registry.yourdomain.com
Setting up Hubble infrastructure...
Starting server on :5000
```

Press `Ctrl+C` to exit logs.

### 3.5 Verify Installation

```bash
# Check all containers are running
docker ps

# You should see:
# - hubble-server
# - hubble-traefik
# - hubble-registry
```

**Congratulations!** Hubble is now running! 🎉

---

## Part 4: First Login (2 minutes)

### 4.1 Access Hubble API

```bash
# From your local machine or server
curl http://localhost:3000/

# Should return: hubble
```

### 4.2 Login to Get Cookies

```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"YourSecurePassword123!"}' \
  -c ~/hubble-cookies.txt \
  -v
```

Look for: `"authenticated": true`

### 4.3 Test Authentication

```bash
curl http://localhost:3000/auth/me \
  -b ~/hubble-cookies.txt

# Should return your username
```

---

## Part 5: Login to Registry (3 minutes)

### 5.1 Wait for HTTPS Certificate

It takes 1-2 minutes for Let's Encrypt to issue certificates. Check status:

```bash
docker logs hubble-traefik | grep -i certificate

# Look for: "Certificate obtained for domain registry.yourdomain.com"
```

### 5.2 Login to Registry

```bash
docker login registry.yourdomain.com
Username: admin
Password: YourSecurePassword123!
```

Expected output:
```
Login Succeeded
```

### 5.3 Test Registry Access

```bash
# List catalog (should be empty for now)
curl https://registry.yourdomain.com/v2/_catalog \
  -u admin:YourSecurePassword123!

# Should return: {"repositories":[]}
```

---

## Part 6: Build and Push Your First Image (5 minutes)

### 6.1 Create a Simple App

```bash
mkdir ~/test-app
cd ~/test-app

# Create a simple HTML file
cat > index.html <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>My Hubble App</title>
</head>
<body>
    <h1>Hello from Hubble!</h1>
    <p>This app is running on my own infrastructure!</p>
</body>
</html>
EOF

# Create Dockerfile
cat > Dockerfile <<'EOF'
FROM nginx:alpine
COPY index.html /usr/share/nginx/html/index.html
EXPOSE 80
EOF
```

### 6.2 Build Image

```bash
docker build -t myapp:v1.0 .
```

### 6.3 Tag for Your Registry

```bash
docker tag myapp:v1.0 registry.yourdomain.com/myapp:v1.0
```

### 6.4 Push to Your Registry

```bash
docker push registry.yourdomain.com/myapp:v1.0
```

Expected output:
```
The push refers to repository [registry.yourdomain.com/myapp]
...
v1.0: digest: sha256:abc123... size: 1234
```

### 6.5 Verify Image in Registry

```bash
curl https://registry.yourdomain.com/v2/_catalog \
  -u admin:YourSecurePassword123!

# Should return: {"repositories":["myapp"]}
```

**Success!** Your first image is in your private registry! 🚀

---

## Part 7: Deploy Your App (5 minutes)

### 7.1 Create Project

```bash
curl -X POST http://localhost:3000/projects \
  -H "Content-Type: application/json" \
  -b ~/hubble-cookies.txt \
  -d '{"name":"myapp"}'
```

### 7.2 Add Hubble Network

```bash
curl -X POST http://localhost:3000/projects/myapp/networks \
  -H "Content-Type: application/json" \
  -b ~/hubble-cookies.txt \
  -d '{"name":"hubble","external":true}'
```

### 7.3 Add Web Service

```bash
curl -X POST http://localhost:3000/projects/myapp/services \
  -H "Content-Type: application/json" \
  -b ~/hubble-cookies.txt \
  -d '{
    "name": "web",
    "image": "registry.yourdomain.com/myapp:v1.0",
    "networks": ["hubble"],
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.myapp.rule=Host(`myapp.yourdomain.com`)",
      "traefik.http.routers.myapp.entrypoints=websecure",
      "traefik.http.routers.myapp.tls.certresolver=letsencrypt",
      "traefik.http.services.myapp.loadbalancer.server.port=80"
    ],
    "restart": "unless-stopped"
  }'
```

### 7.4 Start Service

```bash
curl -X POST http://localhost:3000/projects/myapp/services/web/start \
  -b ~/hubble-cookies.txt
```

### 7.5 Check Service Status

```bash
curl http://localhost:3000/projects/myapp/containers \
  -b ~/hubble-cookies.txt

# Should show container running
```

---

## Part 8: Access Your App (2 minutes)

### 8.1 Wait for HTTPS Certificate

```bash
# Watch Traefik get the certificate
docker logs hubble-traefik -f | grep myapp

# Look for: "Certificate obtained for domain myapp.yourdomain.com"
# Press Ctrl+C when you see it
```

### 8.2 Visit Your App

Open your browser and go to:
```
https://myapp.yourdomain.com
```

You should see:
```
Hello from Hubble!
This app is running on my own infrastructure!
```

**🎉 CONGRATULATIONS! Your app is live with automatic HTTPS!**

---

## Part 9: Make an Update (5 minutes)

Let's update the app and deploy a new version.

### 9.1 Update Your App

```bash
cd ~/test-app

# Update the HTML
cat > index.html <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>My Hubble App v2</title>
    <style>
        body { font-family: Arial; text-align: center; padding: 50px; }
        h1 { color: #2563eb; }
    </style>
</head>
<body>
    <h1>Hello from Hubble! 🚀</h1>
    <p>This is version 2.0!</p>
    <p>Updated and redeployed in seconds!</p>
</body>
</html>
EOF
```

### 9.2 Build New Version

```bash
docker build -t myapp:v2.0 .
```

### 9.3 Push to Registry

```bash
docker tag myapp:v2.0 registry.yourdomain.com/myapp:v2.0
docker push registry.yourdomain.com/myapp:v2.0
```

### 9.4 Update Service

```bash
curl -X PUT http://localhost:3000/projects/myapp/services/web \
  -H "Content-Type: application/json" \
  -b ~/hubble-cookies.txt \
  -d '{
    "image": "registry.yourdomain.com/myapp:v2.0"
  }'
```

### 9.5 Restart Service

```bash
# Stop old version
curl -X POST http://localhost:3000/projects/myapp/services/web/stop \
  -b ~/hubble-cookies.txt

# Start new version
curl -X POST http://localhost:3000/projects/myapp/services/web/start \
  -b ~/hubble-cookies.txt
```

### 9.6 Verify Update

Refresh `https://myapp.yourdomain.com` in your browser.

You should see version 2.0!

---

## What You've Accomplished

In just 30 minutes, you've:

✅ Set up a complete self-hosted Docker platform  
✅ Configured automatic HTTPS with Let's Encrypt  
✅ Created a private Docker registry  
✅ Built and pushed Docker images  
✅ Deployed an app with automatic routing  
✅ Updated and redeployed an application  

**You now have a professional deployment platform!**

---

## Next Steps

### Monitor Your Platform

```bash
# View all containers
docker ps

# View platform logs
docker-compose logs -f

# View Traefik dashboard (SSH tunnel required)
ssh -L 8080:localhost:8080 user@your-server
# Then visit: http://localhost:8080
```

### Deploy More Apps

```bash
# Create another project
curl -X POST http://localhost:3000/projects \
  -b ~/hubble-cookies.txt \
  -d '{"name":"blog"}'

# Follow the same pattern:
# 1. Add network
# 2. Add service
# 3. Start service
```

### Explore the API

See [API.md](API.md) for complete API documentation:
- Project management
- Service configuration
- Network setup
- Container operations

### Add More Features

- **Database projects**: Deploy PostgreSQL, MySQL, etc.
- **Multi-container apps**: Frontend + backend + database
- **Environment variables**: Configure apps via API
- **Volumes**: Persistent data storage

### Backup Your Data

Hubble uses Docker volumes for persistent storage. Backup using these commands:

```bash
# Backup registry data
docker run --rm -v hubble-registry-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/hubble-registry-backup.tar.gz -C /data .

# Backup registry auth
docker run --rm -v hubble-registry-auth:/data -v $(pwd):/backup alpine \
  tar czf /backup/hubble-registry-auth-backup.tar.gz -C /data .

# Backup Traefik data (certificates)
docker run --rm -v hubble-traefik-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/hubble-traefik-backup.tar.gz -C /data .

# Backup projects
docker run --rm -v hubble-projects:/data -v $(pwd):/backup alpine \
  tar czf /backup/hubble-projects-backup.tar.gz -C /data .

# Backup environment
cp .env .env.backup
```

**Restore from backup:**

```bash
# Restore registry data
docker run --rm -v hubble-registry-data:/data -v $(pwd):/backup alpine \
  tar xzf /backup/hubble-registry-backup.tar.gz -C /data

# Restore other volumes similarly
```

---

## Troubleshooting

### App not accessible

**Problem**: Can't access `myapp.yourdomain.com`

**Solutions**:
1. Check DNS propagation: `dig +short myapp.yourdomain.com`
2. Check container is running: `docker ps | grep myapp`
3. Check Traefik logs: `docker logs hubble-traefik`
4. Wait for certificate (1-2 minutes)

### Registry login fails

**Problem**: `docker login` fails

**Solutions**:
1. Verify DNS: `dig +short registry.yourdomain.com`
2. Check registry is running: `docker ps | grep registry`
3. Check credentials match .env file
4. Ensure HTTPS is working (port 443 open)

### Traefik not starting

**Problem**: Traefik container not running

**Solutions**:
1. Check if port 80/443 are in use: `sudo lsof -i :80`
2. Stop conflicting services: `sudo systemctl stop nginx`
3. Check logs: `docker logs hubble-traefik`

### Build Fails

**Problem**: `docker build` fails

**Solutions**:
1. Check Dockerfile syntax
2. Ensure Docker is running: `docker ps`
3. Check disk space: `df -h`

---

## Get Help

- **Documentation**: [README.md](README.md), [SETUP.md](SETUP.md), [API.md](API.md)
- **Issues**: https://github.com/noel-vega/hubble/issues
- **Community**: https://github.com/noel-vega/hubble/discussions

---

## Congratulations!

You're now running your own self-hosted Docker platform with private registry, automatic HTTPS, and API-driven deployments.

**Welcome to the Hubble community!** 🎉

