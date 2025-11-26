# Traefik Integration

Guide to using Traefik with Hubble for automatic HTTPS and routing.

## Table of Contents

- [What is Traefik?](#what-is-traefik)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Service Examples](#service-examples)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

## What is Traefik?

Traefik is an automatic reverse proxy and load balancer that:
- **Routes traffic** based on hostnames and paths
- **Issues HTTPS certificates** automatically via Let's Encrypt
- **Discovers services** by watching Docker labels
- **Provides a dashboard** for monitoring

When Traefik is enabled in Hubble, it auto-provisions on startup and connects to the `hubble` network.

**Platform Services with Traefik:**
- `registry.yourdomain.com` → Hubble Registry (auto-configured)
- `api.yourdomain.com` → Hubble API (optional)
- `*.yourdomain.com` → Your deployed apps

## Quick Start

### 1. Enable Traefik

```bash
# In .env file
HUBBLE_DOMAIN=yourdomain.com
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
```

### 2. Restart Hubble

```bash
docker-compose down
docker-compose up -d
```

You should see:
```
✓ Hubble network 'hubble' created
Creating Traefik container...
✓ Created and started Traefik (ID: abc123)
  HTTP: http://0.0.0.0:80
  HTTPS: https://0.0.0.0:443
✓ Registry started
  Registry URL: https://registry.yourdomain.com
```

The registry is now accessible with automatic HTTPS!

### 3. Add a Service with Traefik Labels

```bash
curl -X POST http://localhost:3000/projects/blog/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "web",
    "image": "nginx:alpine",
    "networks": ["hubble"],
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.blog.rule=Host(`blog.yourdomain.com`)",
      "traefik.http.routers.blog.entrypoints=websecure",
      "traefik.http.routers.blog.tls.certresolver=letsencrypt",
      "traefik.http.services.blog.loadbalancer.server.port=80"
    ],
    "restart": "unless-stopped"
  }'
```

### 4. Visit Your Site

```
https://blog.yourdomain.com
```

Traefik will automatically:
- ✅ Route traffic to your service
- ✅ Request an HTTPS certificate from Let's Encrypt
- ✅ Redirect HTTP → HTTPS

## Configuration

### Environment Variables

```bash
# Enable/disable Traefik
HUBBLE_TRAEFIK_ENABLED=true

# Email for Let's Encrypt (required for HTTPS)
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com

# Enable dashboard on localhost:8080
HUBBLE_TRAEFIK_DASHBOARD=true

# Dashboard authentication (htpasswd format)
HUBBLE_TRAEFIK_DASHBOARD_AUTH=$(htpasswd -nb admin yourpassword)
```

### Traefik Ports

When enabled, Traefik listens on:
- **80** - HTTP (auto-redirects to HTTPS)
- **443** - HTTPS
- **8080** - Dashboard (localhost only)

### Accessing the Dashboard

```bash
# If on remote server, create SSH tunnel:
ssh -L 8080:localhost:8080 user@your-server

# Then visit:
http://localhost:8080
```

## Service Examples

### Basic HTTP Service

```json
{
  "name": "app",
  "image": "myapp:latest",
  "networks": ["hubble"],
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.app.rule=Host(`app.example.com`)",
    "traefik.http.services.app.loadbalancer.server.port=3000"
  ]
}
```

### HTTPS with Let's Encrypt

```json
{
  "name": "app",
  "image": "myapp:latest",
  "networks": ["hubble"],
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.app.rule=Host(`app.example.com`)",
    "traefik.http.routers.app.entrypoints=websecure",
    "traefik.http.routers.app.tls.certresolver=letsencrypt",
    "traefik.http.services.app.loadbalancer.server.port=3000"
  ]
}
```

### Multiple Domains

```json
{
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.app.rule=Host(`app.example.com`) || Host(`www.example.com`)",
    "traefik.http.routers.app.entrypoints=websecure",
    "traefik.http.routers.app.tls.certresolver=letsencrypt",
    "traefik.http.services.app.loadbalancer.server.port=3000"
  ]
}
```

### Path-Based Routing

```json
{
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.api.rule=Host(`example.com`) && PathPrefix(`/api`)",
    "traefik.http.routers.api.entrypoints=websecure",
    "traefik.http.routers.api.tls.certresolver=letsencrypt",
    "traefik.http.services.api.loadbalancer.server.port=8080"
  ]
}
```

### Subdomain Routing

```json
{
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.blog.rule=Host(`blog.example.com`)",
    "traefik.http.routers.blog.entrypoints=websecure",
    "traefik.http.routers.blog.tls.certresolver=letsencrypt",
    "traefik.http.services.blog.loadbalancer.server.port=80"
  ]
}
```

### Non-standard Port

```json
{
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.custom.rule=Host(`custom.example.com`)",
    "traefik.http.routers.custom.entrypoints=websecure",
    "traefik.http.routers.custom.tls.certresolver=letsencrypt",
    "traefik.http.services.custom.loadbalancer.server.port=8888"
  ]
}
```

## Common Patterns

### WordPress Site

```bash
curl -X POST http://localhost:3000/projects/wordpress/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "wordpress",
    "image": "wordpress:latest",
    "environment": {
      "WORDPRESS_DB_HOST": "db",
      "WORDPRESS_DB_PASSWORD": "secret"
    },
    "networks": ["hubble"],
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.wp.rule=Host(`myblog.com`)",
      "traefik.http.routers.wp.entrypoints=websecure",
      "traefik.http.routers.wp.tls.certresolver=letsencrypt",
      "traefik.http.services.wp.loadbalancer.server.port=80"
    ],
    "restart": "unless-stopped"
  }'
```

### Next.js App

```bash
curl -X POST http://localhost:3000/projects/nextjs/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "app",
    "image": "myregistry.com/nextjs-app:latest",
    "networks": ["hubble"],
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.nextjs.rule=Host(`app.example.com`)",
      "traefik.http.routers.nextjs.entrypoints=websecure",
      "traefik.http.routers.nextjs.tls.certresolver=letsencrypt",
      "traefik.http.services.nextjs.loadbalancer.server.port=3000"
    ],
    "restart": "unless-stopped"
  }'
```

### API with Multiple Endpoints

```bash
# Main API
curl -X POST http://localhost:3000/projects/api/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "api",
    "image": "myapi:latest",
    "networks": ["hubble"],
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.api.rule=Host(`api.example.com`)",
      "traefik.http.routers.api.entrypoints=websecure",
      "traefik.http.routers.api.tls.certresolver=letsencrypt",
      "traefik.http.services.api.loadbalancer.server.port=8080"
    ]
  }'

# Admin panel (same app, different router)
curl -X PUT http://localhost:3000/projects/api/services/api \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "labels": [
      "traefik.enable=true",
      "traefik.http.routers.api.rule=Host(`api.example.com`)",
      "traefik.http.routers.api.entrypoints=websecure",
      "traefik.http.routers.api.tls.certresolver=letsencrypt",
      "traefik.http.routers.admin.rule=Host(`admin.example.com`)",
      "traefik.http.routers.admin.entrypoints=websecure",
      "traefik.http.routers.admin.tls.certresolver=letsencrypt",
      "traefik.http.services.api.loadbalancer.server.port=8080"
    ]
  }'
```

## Label Reference

### Required Labels

```
traefik.enable=true
  Enable Traefik for this service

traefik.http.routers.<name>.rule=Host(`example.com`)
  Routing rule (hostname, path, etc.)

traefik.http.services.<name>.loadbalancer.server.port=80
  Container port to route traffic to
```

### HTTPS Labels

```
traefik.http.routers.<name>.entrypoints=websecure
  Use HTTPS entrypoint (port 443)

traefik.http.routers.<name>.tls.certresolver=letsencrypt
  Use Let's Encrypt for certificates
```

### Routing Rules

```
Host(`example.com`)
  Match exact hostname

Host(`example.com`) || Host(`www.example.com`)
  Match multiple hostnames

Host(`example.com`) && PathPrefix(`/api`)
  Match hostname + path

PathPrefix(`/api`)
  Match any request starting with /api

Path(`/health`)
  Match exact path
```

## Troubleshooting

### Certificate Not Issued

**Problem**: Site shows "certificate error" or uses self-signed cert

**Solutions**:

1. Check Let's Encrypt rate limits:
   - 50 certificates per domain per week
   - 5 duplicate certificates per week
   - Visit: https://letsencrypt.org/docs/rate-limits/

2. Verify DNS is pointing to your server:
   ```bash
   dig +short yourdomain.com
   # Should show your server's IP
   ```

3. Check ports 80/443 are accessible:
   ```bash
   curl -I http://yourdomain.com
   curl -I https://yourdomain.com
   ```

4. Review Traefik logs:
   ```bash
   docker logs hubble-traefik
   ```

### Service Not Routing

**Problem**: Getting 404 or "Service Unavailable"

**Solutions**:

1. Verify service is on `hubble` network:
   ```bash
   docker inspect <container-name> | grep -A 10 Networks
   ```

2. Check `traefik.enable=true` label is set:
   ```bash
   docker inspect <container-name> | grep traefik.enable
   ```

3. Verify port is correct:
   ```bash
   # Check what port your app actually listens on
   docker logs <container-name>
   ```

4. Check Traefik dashboard:
   ```
   http://localhost:8080
   ```
   Look under "HTTP Routers" and "HTTP Services"

### HTTP Not Redirecting to HTTPS

**Problem**: HTTP traffic not redirecting to HTTPS

**Check Traefik configuration**:

The HTTP → HTTPS redirect is enabled by default in Hubble's Traefik setup. If it's not working:

1. Verify Traefik is using the correct entrypoints:
   ```bash
   docker logs hubble-traefik | grep entrypoint
   ```

2. Ensure service uses `websecure` entrypoint:
   ```
   traefik.http.routers.app.entrypoints=websecure
   ```

### Dashboard Not Accessible

**Problem**: Cannot access Traefik dashboard

**Solutions**:

1. Verify dashboard is enabled:
   ```bash
   echo $HUBBLE_TRAEFIK_DASHBOARD
   # Should be 'true'
   ```

2. Dashboard is localhost-only. If on remote server:
   ```bash
   ssh -L 8080:localhost:8080 user@your-server
   ```

3. Check Traefik is running:
   ```bash
   docker ps | grep traefik
   ```

### Port Conflict

**Problem**: Traefik won't start - ports 80/443 in use

**Solutions**:

1. Check what's using the ports:
   ```bash
   sudo lsof -i :80
   sudo lsof -i :443
   ```

2. Stop conflicting services:
   ```bash
   sudo systemctl stop nginx
   sudo systemctl stop apache2
   ```

3. Or disable Traefik and use direct port mapping

## DNS Configuration

For Traefik to work properly, configure your DNS:

### Single Domain

```
Type    Name    Value           TTL
A       @       YOUR.SERVER.IP  300
A       www     YOUR.SERVER.IP  300
```

### Wildcard (Multiple Subdomains)

```
Type    Name    Value           TTL
A       @       YOUR.SERVER.IP  300
A       *       YOUR.SERVER.IP  300
```

This allows:
- `yourdomain.com` → Your site
- `blog.yourdomain.com` → Your blog
- `api.yourdomain.com` → Your API
- etc.

## Best Practices

1. **Always use HTTPS in production**
   - Set `entrypoints=websecure`
   - Set `tls.certresolver=letsencrypt`

2. **Use meaningful router names**
   - Good: `traefik.http.routers.blog-web...`
   - Bad: `traefik.http.routers.router1...`

3. **Connect to hubble network**
   - All Traefik-enabled services must be on `hubble` network

4. **Specify the correct port**
   - Check your app's documentation
   - Common: 80 (nginx), 3000 (Node.js), 8080 (Java)

5. **Monitor certificate expiration**
   - Let's Encrypt certs auto-renew
   - Check dashboard periodically

6. **Use path prefixes for APIs**
   - `/api` → API service
   - `/admin` → Admin panel
   - Cleaner than multiple domains

## Resources

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Docker Labels Reference](https://doc.traefik.io/traefik/routing/providers/docker/)
- [Let's Encrypt](https://letsencrypt.org/)
- [Hubble API Reference](API.md)
