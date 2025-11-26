# HTTPS Configuration Summary

## Problem
Browser showed "Site not secure" when accessing Hubble.

## Root Cause
1. `HUBBLE_DOMAIN=localhost` - Let's Encrypt cannot issue certificates for localhost
2. `HUBBLE_TRAEFIK_EMAIL` not set - Required for Let's Encrypt registration
3. All services configured for HTTPS (`websecure` entrypoint) but no certificates

## Solution

### For Development (localhost) - HTTP Only
**File: `docker-compose.yml`**
- Changed all router entrypoints from `websecure` to `web`
- Removed `tls.certresolver=letsencrypt` labels
- Commented out HTTPS redirect in Traefik
- Services now accessible via HTTP:
  - http://hubble.localhost (Web UI)
  - http://hubble.localhost/api (API)
  - http://registry.localhost (Registry)

### For Production (real domain) - HTTPS with Let's Encrypt
**File: `docker-compose.prod.yml`** (new file)
- Override file for production deployment
- Enables HTTPS redirect
- Configures Let's Encrypt certificate resolver
- Sets all routers to `websecure` entrypoint with TLS
- Requires:
  - Real domain (e.g., `yourdomain.com`)
  - `HUBBLE_TRAEFIK_EMAIL` set in `.env`
  - DNS records pointing to server
  - Ports 80 and 443 open

## Usage

### Development
```bash
# .env
HUBBLE_DOMAIN=localhost
HUBBLE_TRAEFIK_EMAIL=  # Leave empty

# Start
docker compose up -d

# Access
open http://hubble.localhost
```

### Production
```bash
# .env
HUBBLE_DOMAIN=yourdomain.com
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com

# Start with production override
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Access
open https://hubble.yourdomain.com
```

## Files Changed
- `docker-compose.yml` - Default to HTTP for development
- `docker-compose.prod.yml` - NEW: Production override with HTTPS
- `.env.example` - Updated with clearer instructions
- `DEPLOYMENT_GUIDE.md` - NEW: Comprehensive deployment guide

## Testing
✅ HTTP access works: `curl http://hubble.localhost`
✅ No "site not secure" warning (using HTTP intentionally)
✅ All services start successfully
✅ Ready for production deployment with real domain
