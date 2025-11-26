# Hubble Deployment Guide

This guide explains how to deploy Hubble in different environments with proper HTTPS configuration.

## Table of Contents
- [Development (localhost)](#development-localhost)
- [Production (with HTTPS)](#production-with-https)
- [Troubleshooting](#troubleshooting)

---

## Development (localhost)

For local development without HTTPS (HTTP only).

### 1. Configure Environment

```bash
cp .env.example .env
```

Edit `.env`:
```bash
ENVIRONMENT=development
ADMIN_USERNAME=admin
ADMIN_PASSWORD=testpass123
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# Use localhost - no HTTPS
HUBBLE_DOMAIN=localhost
# Leave email empty for development
HUBBLE_TRAEFIK_EMAIL=
```

### 2. Start Services

```bash
docker compose up -d
```

### 3. Access Services

- **Web UI:** http://hubble.localhost
- **API:** http://hubble.localhost/api
- **Registry:** http://registry.localhost


### 4. Add to /etc/hosts (if needed)

If `*.localhost` doesn't work, add to `/etc/hosts`:
```
127.0.0.1 hubble.localhost
127.0.0.1 registry.localhost
```

**Note:** Browsers will show "Not Secure" because there's no HTTPS. This is expected for local development.

---

## Production (with HTTPS)

For production deployment with automatic HTTPS via Let's Encrypt.

### Prerequisites

1. **Public domain** pointing to your server (e.g., `yourdomain.com`)
2. **DNS records** configured:
   ```
   A    hubble.yourdomain.com    → your-server-ip
   A    registry.yourdomain.com  → your-server-ip
   ```
3. **Ports open:** 80 (HTTP) and 443 (HTTPS)
4. **Valid email** for Let's Encrypt notifications

### 1. Configure Environment

```bash
cp .env.example .env
```

Edit `.env`:
```bash
ENVIRONMENT=production
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password-here
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# Your actual domain
HUBBLE_DOMAIN=yourdomain.com

# Required for Let's Encrypt
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
```

### 2. Start Services (Production Mode)

```bash
# Use production override for HTTPS
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 3. Verify HTTPS Certificates

Watch Traefik logs to see Let's Encrypt certificate generation:
```bash
docker compose logs -f hubble-traefik
```

You should see:
```
[acme] Trying to solve DNS-01
[acme] The server validated our request
[acme] Certificate obtained successfully
```

### 4. Access Services (HTTPS)

- **Web UI:** https://hubble.yourdomain.com
- **API:** https://hubble.yourdomain.com/api
- **Registry:** https://registry.yourdomain.com


### 5. Test Docker Registry

Login to your private registry:
```bash
docker login registry.yourdomain.com -u admin -p your-password
```

Push an image:
```bash
docker tag myapp:latest registry.yourdomain.com/myapp:latest
docker push registry.yourdomain.com/myapp:latest
```

---

## Configuration Comparison

| Feature | Development | Production |
|---------|-------------|------------|
| Domain | `localhost` | `yourdomain.com` |
| Protocol | HTTP (port 80) | HTTPS (port 443) |
| Certificates | None | Let's Encrypt |
| HTTPS Redirect | Disabled | Enabled |
| Email Required | No | Yes |
| Secure Cookies | No | Yes |
| Docker Compose | `docker compose up -d` | `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d` |

---

## Troubleshooting

### "Site Not Secure" in Browser

**Development (localhost):**
- Expected behavior! Let's Encrypt doesn't issue certificates for localhost
- Access via HTTP: `http://hubble.localhost`
- Browser will show "Not Secure" - this is normal

**Production (real domain):**
- Check DNS: `dig hubble.yourdomain.com` should return your server IP
- Check ports: `netstat -tulpn | grep -E ':(80|443)'`
- Check Traefik logs: `docker compose logs hubble-traefik`
- Verify email is set: `grep HUBBLE_TRAEFIK_EMAIL .env`

### Let's Encrypt Rate Limits

If testing repeatedly, you may hit rate limits (5 certificates per week per domain). Use Let's Encrypt staging server for testing:

Add to `docker-compose.prod.yml`:
```yaml
services:
  hubble-traefik:
    command:
      # ... existing commands ...
      - --certificatesresolvers.letsencrypt.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory
```

Remove this line once everything works!

### Certificate Not Generating

1. **Check DNS propagation:**
   ```bash
   dig hubble.yourdomain.com
   dig registry.yourdomain.com
   ```

2. **Check Traefik can reach Let's Encrypt:**
   ```bash
   docker compose exec hubble-traefik wget -O- https://acme-v02.api.letsencrypt.org/directory
   ```

3. **Check ports are accessible:**
   ```bash
   # From external machine
   curl -I http://yourdomain.com
   ```

4. **Check email is set:**
   ```bash
   docker compose exec hubble-traefik env | grep EMAIL
   ```

### "Connection Refused" Error

- Check all services are running: `docker compose ps`
- Check Traefik routes: `http://localhost:8080` (dashboard)
- Check service labels: `docker inspect hubble-server | grep traefik`

### Registry Push Fails

1. **Check you're logged in:**
   ```bash
   docker login registry.yourdomain.com
   ```

2. **Check registry is accessible:**
   ```bash
   curl -u admin:password https://registry.yourdomain.com/v2/_catalog
   ```

3. **Check firewall/DNS:**
   ```bash
   ping registry.yourdomain.com
   ```

---

## Switching Between Environments

### Development → Production

```bash
# Stop development stack
docker compose down

# Update .env with production values
vim .env

# Start production stack
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Production → Development

```bash
# Stop production stack
docker compose -f docker-compose.yml -f docker-compose.prod.yml down

# Update .env with development values
vim .env

# Start development stack
docker compose up -d
```

---

## Best Practices

### Development
- ✅ Use `HUBBLE_DOMAIN=localhost`
- ✅ Leave `HUBBLE_TRAEFIK_EMAIL` empty
- ✅ Access via HTTP (`http://`)
- ✅ Use simple passwords for testing

### Production
- ✅ Use real domain (e.g., `HUBBLE_DOMAIN=yourdomain.com`)
- ✅ Set valid email for Let's Encrypt notifications
- ✅ Use strong passwords (min 16 chars)
- ✅ Use strong JWT secrets (`openssl rand -base64 32`)
- ✅ Set `ENVIRONMENT=production` for secure cookies
- ✅ Enable firewall (allow 80, 443, 22 only)
- ✅ Keep `.env` secure (never commit to git)
- ✅ Backup volumes regularly

---

## Quick Reference

### Development Commands
```bash
# Start
docker compose up -d

# Logs
docker compose logs -f

# Stop
docker compose down

# Access
open http://hubble.localhost
```

### Production Commands
```bash
# Start
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Logs
docker compose logs -f

# Stop  
docker compose -f docker-compose.yml -f docker-compose.prod.yml down

# Access
open https://hubble.yourdomain.com
```

### Useful Commands
```bash
# Check certificates
docker compose exec hubble-traefik cat /data/acme.json | jq

# Test registry
curl -u admin:pass https://registry.yourdomain.com/v2/_catalog

# View Traefik config
curl http://localhost:8080/api/rawdata

# Restart single service
docker compose restart hubble-server
```
