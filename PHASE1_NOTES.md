# Phase 1: Infrastructure in docker-compose.yml

## What Changed

Traefik and Registry are now defined in `docker-compose.yml` alongside hubble-server and hubble-web.

### Benefits:
- ✅ All services visible in one file
- ✅ Standard Docker Compose workflow
- ✅ Easier debugging and configuration
- ✅ Faster startup (parallel container creation)
- ✅ Better for production deployments

### Backward Compatibility:
The Go code in `platform/traefik.go` and `platform/registry.go` still exists as a **fallback**. If the containers don't exist, Hubble will create them programmatically.

## docker-compose.yml Services:

```yaml
services:
  hubble-server:    # API backend
  hubble-web:       # React frontend
  hubble-traefik:   # Reverse proxy (NEW!)
  registry-init:    # One-time auth setup (NEW!)
  hubble-registry:  # Docker registry (NEW!)
```

## Usage:

### Start Everything:
```bash
docker compose up -d
```

This starts:
1. **hubble-traefik** - Handles HTTPS and routing
2. **registry-init** - Creates htpasswd file (runs once and exits)
3. **hubble-registry** - Docker image storage
4. **hubble-server** - API backend (detects existing Traefik/Registry, doesn't recreate)
5. **hubble-web** - React frontend

### Configuration:

All configuration via environment variables in `.env`:

```bash
# Traefik HTTPS
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com

# Registry auth (uses same admin credentials)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=yourpassword
```

## Migration Path:

**Phase 1** (Current): Infrastructure in compose + Go fallback  
**Phase 2** (Future): Make Go code optional  
**Phase 3** (Future): Remove platform/* code entirely (v2.0)

This gives users time to adapt while maintaining backward compatibility.
