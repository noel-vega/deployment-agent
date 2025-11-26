# Phase 3: Complete Removal of Programmatic Infrastructure Creation

## What Changed

🎉 **Hubble v2.0** - Removed ~500 lines of container creation code! Infrastructure is now **100% declarative** via docker-compose.yml.

### Code Changes:

**Before (v1.x):**
- `platform/traefik.go`: 272 lines (container creation, config management, image pulling)
- `platform/registry.go`: 273 lines (container creation, config management, htpasswd creation)
- Total: ~545 lines of complex Docker SDK code

**After (v2.0):**
- `platform/traefik.go`: 58 lines (-79% reduction)
- `platform/registry.go`: 57 lines (-79% reduction)
- Total: ~115 lines (simple status checks only)

### Removed Functions:
- ❌ `createTraefikContainer()` - Removed
- ❌ `createRegistryContainer()` - Removed
- ❌ `ensureImageAvailable()` - Removed
- ❌ `GetTraefikConfig()` / `TraefikConfig` struct - Removed
- ❌ `GetRegistryConfig()` / `RegistryConfig` struct - Removed

### Simplified Functions:
- ✅ `EnsureTraefik()` - Now just checks if container exists/running
- ✅ `EnsureRegistry()` - Now just checks if container exists/running
- ✅ `EnsureInfrastructure()` - Logs warnings instead of errors for missing containers

### Removed Environment Variables:
- ❌ `HUBBLE_TRAEFIK_ENABLED` - No longer needed (always enabled via compose)
- ❌ `HUBBLE_REGISTRY_ENABLED` - No longer needed (always enabled via compose)
- ❌ `HUBBLE_DISABLE_PLATFORM_FALLBACK` - No longer needed (no fallback exists)

## New Behavior

### When All Services Running (Normal):
```
✓ Hubble network 'hubble' already exists
✓ Traefik running (ID: abc123)
✓ Registry running (ID: def456)
Starting server on :5000
```

### When Infrastructure Missing:
```
⚠️  Traefik container not found!
⚠️  Please ensure hubble-traefik service is running via docker-compose.yml
⚠️  Run: docker compose up -d
Warning: Traefik container not found - start via docker-compose

⚠️  Registry container not found!
⚠️  Please ensure hubble-registry service is running via docker-compose.yml
⚠️  Run: docker compose up -d
Warning: Registry container not found - start via docker-compose

Starting server on :5000
```

**Note:** Server still starts, but infrastructure features won't work. Just run `docker compose up -d` to fix.

## Why This Is Better

### Before (v1.x - Programmatic):
```go
// 272 lines in platform/traefik.go
func createTraefikContainer(cli *client.Client, config TraefikConfig) error {
    // Pull image
    // Create container with 50+ config options
    // Handle volume mounts
    // Configure labels
    // Start container
    // etc...
}
```

### After (v2.0 - Declarative):
```yaml
# docker-compose.yml
services:
  hubble-traefik:
    image: traefik:v3.0
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    command:
      - --providers.docker=true
      # ...
```

```go
// 58 lines in platform/traefik.go
func EnsureTraefik(cli *client.Client) error {
    // Just check if container exists/running
    // Log warnings if missing
    // That's it!
}
```

### Benefits:
✅ **Simpler codebase** - 79% less code to maintain  
✅ **All infrastructure visible** - No hidden containers  
✅ **Standard Docker workflow** - Industry best practice  
✅ **Easier debugging** - `docker compose logs`, `docker compose ps`  
✅ **Better startup** - Parallel container creation  
✅ **Git-trackable config** - Infrastructure is code  
✅ **No Docker SDK complexity** - Just YAML  

## Migration from v1.x

### Breaking Changes:
- ⚠️ **Requires docker-compose.yml** - Programmatic creation removed
- ⚠️ **Environment variables removed** - See above

### How to Migrate:

**Step 1:** Ensure you have the v2.0 docker-compose.yml:
```bash
# It should include these services:
# - hubble-server
# - hubble-web
# - hubble-traefik
# - registry-init
# - hubble-registry
```

**Step 2:** Update your .env file (remove deprecated variables):
```bash
# REMOVED (no longer needed):
# HUBBLE_TRAEFIK_ENABLED=true
# HUBBLE_REGISTRY_ENABLED=true
# HUBBLE_DISABLE_PLATFORM_FALLBACK=true

# KEEP (still needed):
HUBBLE_DOMAIN=yourdomain.com
HUBBLE_TRAEFIK_EMAIL=admin@yourdomain.com
ADMIN_USERNAME=admin
ADMIN_PASSWORD=yourpassword
```

**Step 3:** Restart services:
```bash
docker compose down
docker compose up -d
```

**Step 4:** Verify:
```bash
docker compose ps
# Should show: hubble-server, hubble-web, hubble-traefik, hubble-registry
```

## Architecture

### v2.0 Service Architecture:
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
└─────────────────────────────────────────────────────────┘
```

### What hubble-server Does in v2.0:
1. Checks if `hubble` network exists (creates if missing)
2. Checks if `hubble-traefik` container exists (warns if missing)
3. Checks if `hubble-registry` container exists (warns if missing)
4. Starts API server

**No container creation. No image pulling. Just status checks.**

## Code Cleanup Summary

| File | Before | After | Removed | Change |
|------|--------|-------|---------|--------|
| `platform/traefik.go` | 272 lines | 58 lines | 214 lines | -79% |
| `platform/registry.go` | 273 lines | 57 lines | 216 lines | -79% |
| `platform/infrastructure.go` | 45 lines | 35 lines | 10 lines | -22% |
| `.env.example` | 35 lines | 28 lines | 7 lines | -20% |
| **Total** | **625 lines** | **178 lines** | **447 lines** | **-71%** |

## Timeline

- ✅ **Phase 0** (v0.x): Host bind mounts → Docker volumes migration
- ✅ **Phase 1** (v1.0): Added infrastructure to docker-compose.yml + Go fallback
- ✅ **Phase 2** (v1.1): Added deprecation warnings for Go fallback
- ✅ **Phase 3** (v2.0): Removed programmatic creation entirely

## Testing

### Test 1: Normal Operation (All Services Running)
```bash
docker compose up -d
docker compose logs hubble-server
# Expected: ✓ Traefik running, ✓ Registry running
```

### Test 2: Missing Infrastructure
```bash
docker rm -f hubble-traefik hubble-registry
docker compose restart hubble-server
docker compose logs hubble-server
# Expected: ⚠️ warnings, server still starts
```

### Test 3: Recovery
```bash
docker compose up -d
# Expected: All services start, warnings disappear
```

## Conclusion

**Hubble v2.0 is a major simplification:**
- Removed 447 lines of code (-71%)
- 100% declarative infrastructure via docker-compose.yml
- Industry-standard Docker workflow
- Easier to understand, debug, and maintain

**No more hidden containers. Everything is visible and git-trackable.**
