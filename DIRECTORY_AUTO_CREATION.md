# Automatic Directory Creation

## Overview

Hubble automatically creates all required storage directories on startup. **No manual directory creation is needed!**

## What Gets Created Automatically

When you run `docker compose up -d`, Hubble will automatically create:

1. **Traefik Data Directory** (`/var/lib/hubble/traefik`)
   - Stores Let's Encrypt certificates (acme.json)
   - Created with 0755 permissions
   - acme.json created with 0600 permissions (security requirement)

2. **Registry Storage Directory** (configurable via `HUBBLE_REGISTRY_STORAGE`)
   - Default: `/var/lib/hubble/registry`
   - Stores all Docker images
   - Created with 0755 permissions

3. **Registry Auth Directory** (derived from storage path)
   - Default: `/var/lib/hubble/registry-auth`
   - Stores htpasswd authentication file
   - Created with 0755 permissions

## How It Works

The directory creation happens in the `platform/` code:

### Traefik (`platform/traefik.go`)

```go
// Automatically creates /var/lib/hubble/traefik
if err := os.MkdirAll(TraefikDataPath, 0755); err != nil {
    return fmt.Errorf("failed to create Traefik data directory: %w", err)
}

// Verifies directory exists
if info, err := os.Stat(TraefikDataPath); err != nil {
    return fmt.Errorf("directory does not exist after creation: %w", err)
}
```

### Registry (`platform/registry.go`)

```go
// Automatically creates storage directory
if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
    return fmt.Errorf("failed to create storage directory: %w", err)
}

// Automatically creates auth directory  
if err := os.MkdirAll(config.AuthPath, 0755); err != nil {
    return fmt.Errorf("failed to create auth directory: %w", err)
}
```

## Logging

You'll see these log messages on startup:

```
Creating Traefik data directory: /var/lib/hubble/traefik
✓ Traefik data directory ready: /var/lib/hubble/traefik
Creating Registry storage directory: /var/lib/hubble/registry
Creating Registry auth directory: /var/lib/hubble/registry-auth
✓ Registry directories ready
```

## Permission Requirements

The hubble-server container runs with **Docker socket access**, which means it can create directories on the host system. However:

### On Linux (Standard Setup)

- Directories are created as the user running Docker
- Usually works without issues
- If `/var/lib` is not writable, you may see permission errors

### Permission Error?

If you see:
```
failed to create Traefik data directory: permission denied
```

**Solution 1: Use sudo to pre-create the parent directory**
```bash
sudo mkdir -p /var/lib/hubble
sudo chown $USER:$USER /var/lib/hubble
```

**Solution 2: Use a different storage path (no sudo needed)**
```env
# In .env file
HUBBLE_REGISTRY_STORAGE=/home/youruser/hubble-data/registry
```

Then Hubble will automatically create:
- `/home/youruser/hubble-data/registry`
- `/home/youruser/hubble-data/registry-auth`

For Traefik, you'd need to modify the code or pre-create `/var/lib/hubble/traefik`.

## Testing Directory Creation

To test if automatic directory creation works:

1. Remove existing directories:
   ```bash
   sudo rm -rf /var/lib/hubble/
   ```

2. Start Hubble:
   ```bash
   docker compose up -d
   ```

3. Check logs:
   ```bash
   docker logs hubble-server
   ```

   You should see:
   ```
   Creating Traefik data directory: /var/lib/hubble/traefik
   ✓ Traefik data directory ready
   Creating Registry storage directory: /var/lib/hubble/registry  
   Creating Registry auth directory: /var/lib/hubble/registry-auth
   ✓ Registry directories ready
   ```

4. Verify directories exist:
   ```bash
   ls -la /var/lib/hubble/
   ```

   Expected output:
   ```
   drwxr-xr-x registry/
   drwxr-xr-x registry-auth/
   drwxr-xr-x traefik/
   ```

## Troubleshooting

### Directories Not Created?

Check the logs for errors:
```bash
docker logs hubble-server | grep -i "failed to create"
```

### Permission Denied?

Pre-create the parent directory with proper ownership:
```bash
sudo mkdir -p /var/lib/hubble
sudo chown $USER:$USER /var/lib/hubble
```

### Still Having Issues?

The container might not have access to the host filesystem. Check:
1. Docker socket is mounted: `-v /var/run/docker.sock:/var/run/docker.sock`
2. User has permissions to write to `/var/lib`

## For Developers

To add automatic directory creation for new components:

```go
func createMyContainer(dockerClient *client.Client, config MyConfig) error {
    // 1. Define the path
    myDataPath := "/var/lib/hubble/mycomponent"
    
    // 2. Create directory
    log.Printf("Creating directory: %s", myDataPath)
    if err := os.MkdirAll(myDataPath, 0755); err != nil {
        return fmt.Errorf("failed to create directory %s: %w", myDataPath, err)
    }
    
    // 3. Verify it exists
    if info, err := os.Stat(myDataPath); err != nil {
        return fmt.Errorf("directory does not exist after creation: %w", err)
    } else if !info.IsDir() {
        return fmt.Errorf("path exists but is not a directory: %s", myDataPath)
    }
    log.Printf("✓ Directory ready: %s", myDataPath)
    
    // ... rest of container creation
}
```

## Summary

✅ **No manual directory creation needed** - Hubble does it automatically  
✅ **Improved logging** - You can see exactly what's being created  
✅ **Better error messages** - Clear indication if something fails  
✅ **Verification** - Ensures directories exist before using them  

Just configure your `.env` and run `docker compose up -d` - that's it!
