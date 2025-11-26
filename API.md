# API Reference

Complete API reference for Hubble. All protected endpoints require authentication via httpOnly cookies.

## Table of Contents

- [Authentication](#authentication)
- [Projects](#projects)
- [Containers](#containers)
- [Images](#images)
- [Registry](#registry)

## Base URL

```
http://localhost:3000
```

## Authentication

JWT-based authentication with httpOnly cookies. Tokens are automatically sent with requests.

###  `POST /auth/login`

Login and receive access/refresh tokens.

**Request:**
```json
{
  "username": "admin",
  "password": "yourpassword"
}
```

**Response (200 OK):**
```json
{
  "authenticated": true,
  "username": "admin"
}
```

Sets cookies:
- `access_token` (5 minutes, httpOnly)
- `refresh_token` (7 days, httpOnly, path=/auth/refresh)

**Example:**
```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpass"}' \
  -c cookies.txt
```

---

### `POST /auth/refresh`

Refresh expired access token using refresh token.

**Response (200 OK):**
```json
{
  "message": "Token refreshed successfully",
  "expires_in": 300
}
```

Issues new access_token and refresh_token (rotation).

**Example:**
```bash
curl -X POST http://localhost:3000/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

---

### `POST /auth/logout`

Logout and clear authentication cookies.

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/auth/logout \
  -b cookies.txt
```

---

### `GET /auth/me`

Get current authenticated user information.

**Response (200 OK):**
```json
{
  "username": "admin",
  "authenticated": true
}
```

**Example:**
```bash
curl http://localhost:3000/auth/me \
  -b cookies.txt
```

---

## Projects

Manage Docker Compose projects via API.

### `POST /projects`

Create a new project with empty docker-compose.yml.

**Request:**
```json
{
  "name": "my-app"
}
```

**Response (201 Created):**
```json
{
  "message": "project created successfully",
  "project": {
    "name": "my-app",
    "path": "/projects/my-app",
    "service_count": 0,
    "containers_running": 0,
    "containers_stopped": 0
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/projects \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"my-app"}'
```

---

### `GET /projects`

List all projects.

**Response (200 OK):**
```json
{
  "projects": [
    {
      "name": "my-app",
      "path": "/projects/my-app",
      "service_count": 2,
      "containers_running": 2,
      "containers_stopped": 0
    }
  ],
  "count": 1
}
```

---

### `GET /projects/{name}`

Get project details.

**Response (200 OK):**
```json
{
  "name": "my-app",
  "path": "/projects/my-app",
  "service_count": 2,
  "containers_running": 2,
  "containers_stopped": 0
}
```

---

### `GET /projects/{name}/compose`

Get raw docker-compose.yml content.

**Response (200 OK):**
```json
{
  "content": "services:\n  web:\n    image: nginx\n..."
}
```

---

### `GET /projects/{name}/services`

List all services in a project.

**Response (200 OK):**
```json
{
  "services": {
    "web": {
      "image": "nginx:alpine",
      "ports": ["80:80"],
      "restart": "unless-stopped"
    }
  },
  "count": 1
}
```

---

### `POST /projects/{name}/services`

Add a service to a project.

**Request:**
```json
{
  "name": "web",
  "image": "nginx:alpine",
  "ports": ["80:80"],
  "environment": {
    "NGINX_HOST": "example.com"
  },
  "volumes": ["./html:/usr/share/nginx/html"],
  "networks": ["hubble"],
  "restart": "unless-stopped",
  "labels": [
    "traefik.enable=true",
    "traefik.http.routers.web.rule=Host(`example.com`)"
  ]
}
```

**Available fields:**
- `name` (required) - Service name
- `image` - Docker image
- `build` - Build context path
- `ports` - Array of port mappings
- `environment` - Environment variables (object)
- `volumes` - Array of volume mounts
- `depends_on` - Array of service dependencies
- `networks` - Array of networks
- `restart` - Restart policy
- `command` - Override command
- `labels` - Array of labels
- `container_name` - Custom container name

**Response (201 Created):**
```json
{
  "message": "service added successfully",
  "project": "my-app",
  "service": "web"
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/projects/my-app/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "web",
    "image": "nginx:alpine",
    "ports": ["80:80"],
    "restart": "unless-stopped"
  }'
```

---

### `PUT /projects/{name}/services/{service}`

Update an existing service.

**Request:**
```json
{
  "image": "nginx:latest",
  "ports": ["8080:80"]
}
```

All fields optional - only include what you want to update.

**Response (200 OK):**
```json
{
  "message": "service updated successfully",
  "project": "my-app",
  "service": "web"
}
```

---

### `DELETE /projects/{name}/services/{service}`

Remove a service from a project.

**Response (200 OK):**
```json
{
  "message": "service deleted successfully",
  "project": "my-app",
  "service": "web"
}
```

---

### `POST /projects/{name}/services/{service}/start`

Start a specific service.

**Response (200 OK):**
```json
{
  "message": "service started successfully",
  "project": "my-app",
  "service": "web"
}
```

---

### `POST /projects/{name}/services/{service}/stop`

Stop a specific service.

**Response (200 OK):**
```json
{
  "message": "service stopped successfully",
  "project": "my-app",
  "service": "web"
}
```

---

### `GET /projects/{name}/networks`

List all networks in a project.

**Response (200 OK):**
```json
{
  "networks": {
    "hubble": {
      "external": true
    },
    "backend": {
      "driver": "bridge"
    }
  },
  "count": 2
}
```

---

### `POST /projects/{name}/networks`

Add a network to a project.

**Request:**
```json
{
  "name": "hubble",
  "external": true
}
```

**Or create a custom network:**
```json
{
  "name": "backend",
  "driver": "bridge"
}
```

**Validation:**
- External networks (`external: true`) cannot specify a `driver`
- Driver is managed by the existing external network

**Response (201 Created):**
```json
{
  "message": "network added successfully",
  "project": "my-app",
  "network": "hubble"
}
```

---

### `PUT /projects/{name}/networks/{network}`

Update a network configuration.

**Request:**
```json
{
  "driver": "overlay"
}
```

**Response (200 OK):**
```json
{
  "message": "network updated successfully",
  "project": "my-app",
  "network": "backend"
}
```

---

### `DELETE /projects/{name}/networks/{network}`

Remove a network from a project.

**Response (200 OK):**
```json
{
  "message": "network deleted successfully",
  "project": "my-app",
  "network": "backend"
}
```

---

### `GET /projects/{name}/containers`

List containers for a project.

**Response (200 OK):**
```json
{
  "containers": [
    {
      "id": "abc123",
      "name": "my-app-web-1",
      "image": "nginx:alpine",
      "state": "running",
      "status": "Up 5 minutes"
    }
  ],
  "count": 1
}
```

---

### `GET /projects/{name}/volumes`

List volumes for a project.

**Response (200 OK):**
```json
{
  "volumes": {
    "db_data": null
  },
  "count": 1
}
```

---

### `GET /projects/{name}/environment`

Get environment variables for a project.

**Response (200 OK):**
```json
{
  "environment": {
    "NODE_ENV": "production",
    "DATABASE_URL": "postgres://..."
  },
  "count": 2
}
```

---

## Containers

Manage Docker containers directly.

### `GET /containers`

List all Docker containers.

**Response (200 OK):**
```json
{
  "containers": [
    {
      "id": "abc123def456",
      "name": "hubble-server",
      "image": "hubble_hubble-server:latest",
      "state": "running",
      "status": "Up 2 hours",
      "ports": ["3000:5000"]
    }
  ],
  "count": 1
}
```

**Example:**
```bash
curl http://localhost:3000/containers \
  -b cookies.txt
```

---

### `GET /containers/{id}`

Get container details.

**Response (200 OK):**
```json
{
  "id": "abc123def456",
  "name": "hubble-server",
  "image": "hubble_hubble-server:latest",
  "state": "running",
  "status": "Up 2 hours",
  "created": "2024-01-01T00:00:00Z",
  "ports": ["3000:5000"],
  "labels": {
    "com.docker.compose.project": "hubble"
  }
}
```

---

### `POST /containers/{id}/start`

Start a stopped container.

**Response (200 OK):**
```json
{
  "message": "container started successfully",
  "id": "abc123"
}
```

---

### `POST /containers/{id}/stop`

Stop a running container.

**Response (200 OK):**
```json
{
  "message": "container stopped successfully",
  "id": "abc123"
}
```

---

## Images

List Docker images.

### `GET /images`

List all Docker images.

**Response (200 OK):**
```json
{
  "images": [
    {
      "id": "sha256:abc123...",
      "tags": ["nginx:alpine"],
      "size": 23500000,
      "created": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

**Example:**
```bash
curl http://localhost:3000/images \
  -b cookies.txt
```

---

## Registry

Browse self-hosted Docker registries (if configured).

### `GET /registry/repositories`

List all repositories in the configured registry.

**Response (200 OK):**
```json
{
  "registry": "http://localhost:5001",
  "repositories": ["myapp", "blog", "api"],
  "count": 3
}
```

---

### `GET /registry/repositories/{name}/tags`

List all tags for a repository.

**Response (200 OK):**
```json
{
  "registry": "http://localhost:5001",
  "repository": "blog",
  "tags": ["latest", "v1.0", "main"],
  "count": 3
}
```

---

### `GET /registry/catalog`

List all repositories with their tags.

**Response (200 OK):**
```json
{
  "registry": "http://localhost:5001",
  "repositories": [
    {
      "name": "blog",
      "tags": ["latest", "main"]
    },
    {
      "name": "api",
      "tags": ["v1.0", "v2.0"]
    }
  ],
  "count": 2
}
```

---

## Error Responses

All endpoints may return these error responses:

### `400 Bad Request`
```json
{
  "error": "Invalid request body"
}
```

### `401 Unauthorized`
```json
{
  "error": "Unauthorized - No access token"
}
```

### `404 Not Found`
```json
{
  "error": "Project not found: my-app"
}
```

### `409 Conflict`
```json
{
  "error": "Project already exists: my-app"
}
```

### `500 Internal Server Error`
```json
{
  "error": "Failed to create project"
}
```

---

## Complete Workflow Example

### 1. Login
```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpass"}' \
  -c cookies.txt
```

### 2. Create Project
```bash
curl -X POST http://localhost:3000/projects \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"blog"}'
```

### 3. Add Hubble Network
```bash
curl -X POST http://localhost:3000/projects/blog/networks \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"hubble","external":true}'
```

### 4. Add Web Service
```bash
curl -X POST http://localhost:3000/projects/blog/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "web",
    "image": "nginx:alpine",
    "ports": ["80:80"],
    "networks": ["hubble"],
    "restart": "unless-stopped"
  }'
```

### 5. Start Service
```bash
curl -X POST http://localhost:3000/projects/blog/services/web/start \
  -b cookies.txt
```

### 6. Check Containers
```bash
curl http://localhost:3000/projects/blog/containers \
  -b cookies.txt
```

### 7. Logout
```bash
curl -X POST http://localhost:3000/auth/logout \
  -b cookies.txt
```

---

## Notes

- **Cookies**: All auth tokens are stored in httpOnly cookies (secure in production)
- **CORS**: Configure your frontend to send `credentials: 'include'`
- **Version**: Docker Compose `version` field is obsolete and intentionally omitted
- **Networks**: The `hubble` network is auto-created by the platform
- **Labels**: Use labels for Traefik configuration (see [TRAEFIK.md](TRAEFIK.md))
