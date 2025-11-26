# Phase 2: Deprecation Warnings for Programmatic Infrastructure

## What Changed

Platform code now logs **deprecation warnings** when creating Traefik or Registry containers programmatically. This encourages users to use docker-compose.yml instead.

### New Behavior:

When `hubble-traefik` or `hubble-registry` containers don't exist:

```
⚠️  WARNING: Registry container not found!
⚠️  DEPRECATED: Programmatic Registry creation is deprecated.
⚠️  Please add hubble-registry service to your docker-compose.yml
⚠️  See: https://github.com/noel-vega/hubble#infrastructure-services
Creating Registry container via fallback (deprecated)...
```

### New Environment Variable:

```bash
# Disable fallback container creation entirely
HUBBLE_DISABLE_PLATFORM_FALLBACK=true
```

When enabled, Hubble will **refuse to start** if Traefik/Registry containers don't exist, forcing you to use docker-compose.yml.

## Why This Change?

**Benefits of docker-compose approach:**
- ✅ All infrastructure visible in one file
- ✅ Standard Docker workflow
- ✅ Easier debugging
- ✅ Faster startup (parallel)
- ✅ Industry standard

**Drawbacks of programmatic approach:**
- ❌ Hidden containers (~500 lines of Go code)
- ❌ Harder to debug
- ❌ Not standard Docker practice
- ❌ Startup race conditions

## Migration Recommendations:

### For New Users:
Use docker-compose.yml (already the default). No action needed.

### For Existing Users:
If you see deprecation warnings:

1. **Option A: Add services to docker-compose.yml** (Recommended)
   ```bash
   # Use the template in the repo
   docker compose up -d
   ```

2. **Option B: Ignore warnings**  
   Fallback still works. Warnings are just informational.

3. **Option C: Force compose-only mode**  
   Set `HUBBLE_DISABLE_PLATFORM_FALLBACK=true` to ensure you're using compose.

## Timeline:

- **Phase 1** (Current - 1.x): Infrastructure in compose + Go fallback
- **Phase 2** (Current - 1.x): Deprecation warnings for Go fallback
- **Phase 3** (Future - 2.0): Remove platform/* code entirely

No breaking changes in 1.x. Users have time to migrate.
