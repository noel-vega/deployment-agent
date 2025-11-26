# Project Management API

## Overview
APIs for creating and managing Docker Compose projects through a UI form workflow.

## Workflow
1. Create a new project (empty docker-compose.yml with just `services: {}`)
2. Add services to the project via UI form
3. Update services as needed
4. Delete services when no longer needed
5. Start/stop services using existing endpoints

**Note:** The `version` field is obsolete in Docker Compose v2 and is intentionally omitted to avoid deprecation warnings.

## Endpoints

### Create Project
Create a new project with an empty docker-compose.yml file.

```http
POST /projects
Content-Type: application/json

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
    "path": "/path/to/projects/my-app",
    "service_count": 0,
    "containers_running": 0,
    "containers_stopped": 0
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or missing name
- `409 Conflict` - Project already exists
- `500 Internal Server Error` - Server error

---

### Add Service to Project
Add a new service to an existing project.

```http
POST /projects/{name}/services
Content-Type: application/json

{
  "name": "web",
  "image": "nginx:latest",
  "ports": ["80:80", "443:443"],
  "environment": {
    "NGINX_HOST": "example.com"
  },
  "volumes": ["./html:/usr/share/nginx/html"],
  "restart": "unless-stopped"
}
```

**Available Fields:**
- `name` (required) - Service name
- `image` - Docker image (e.g., `nginx:latest`)
- `build` - Build context path
- `ports` - Array of port mappings (e.g., `["8080:80"]`)
- `environment` - Environment variables as key-value map
- `volumes` - Array of volume mounts
- `depends_on` - Array of service dependencies
- `networks` - Array of networks
- `restart` - Restart policy
- `command` - Override default command

**Response (201 Created):**
```json
{
  "message": "service added successfully",
  "project": "my-app",
  "service": "web"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or missing service name
- `404 Not Found` - Project not found
- `409 Conflict` - Service already exists
- `500 Internal Server Error` - Server error

---

### Update Service
Update an existing service in a project.

```http
PUT /projects/{name}/services/{service}
Content-Type: application/json

{
  "image": "nginx:alpine",
  "ports": ["8080:80"],
  "environment": {
    "NGINX_HOST": "newhost.com",
    "DEBUG": "true"
  }
}
```

**Note:** The service name comes from the URL. All fields are optional - only provide fields you want to update.

**Response (200 OK):**
```json
{
  "message": "service updated successfully",
  "project": "my-app",
  "service": "web"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request
- `404 Not Found` - Project or service not found
- `500 Internal Server Error` - Server error

---

### Delete Service
Remove a service from a project.

```http
DELETE /projects/{name}/services/{service}
```

**Response (200 OK):**
```json
{
  "message": "service deleted successfully",
  "project": "my-app",
  "service": "web"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request
- `404 Not Found` - Project or service not found
- `500 Internal Server Error` - Server error

---

## Example: Complete Workflow

### 1. Create Project
```bash
curl -X POST http://localhost:5000/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"name": "my-app"}'
```

### 2. Add Database Service
```bash
curl -X POST http://localhost:5000/projects/my-app/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "db",
    "image": "postgres:14",
    "environment": {
      "POSTGRES_PASSWORD": "secret",
      "POSTGRES_DB": "myapp"
    },
    "volumes": ["db_data:/var/lib/postgresql/data"],
    "restart": "unless-stopped"
  }'
```

### 3. Add Web Service
```bash
curl -X POST http://localhost:5000/projects/my-app/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "web",
    "image": "nginx:latest",
    "ports": ["80:80"],
    "depends_on": ["db"],
    "restart": "unless-stopped"
  }'
```

### 4. Update Web Service
```bash
curl -X PUT http://localhost:5000/projects/my-app/services/web \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "image": "nginx:alpine",
    "ports": ["8080:80"]
  }'
```

### 5. Start Service
```bash
curl -X POST http://localhost:5000/projects/my-app/services/web/start \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 6. View Services
```bash
curl http://localhost:5000/projects/my-app/services \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 7. Delete Service
```bash
curl -X DELETE http://localhost:5000/projects/my-app/services/web \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Technical Notes

### docker-compose.yml Structure
The API maintains a valid docker-compose.yml file with the following structure:
```yaml
services:
  web:
    image: nginx:latest
    ports:
      - 80:80
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: secret
networks: {}  # Preserved if exists
volumes: {}   # Preserved if exists
```

**Note:** The `version` field is obsolete in Docker Compose v2 (as of 2020+) and is intentionally omitted to avoid deprecation warnings.

### What's Preserved
When updating services, the following top-level keys are preserved:
- `services` (required)
- `networks`
- `volumes`
- `secrets`
- `configs`

**Note:** The `version` field is preserved if it exists in existing files (for backward compatibility), but is not added to new projects as it's obsolete in Docker Compose v2.

### What's Not Preserved
- Comments in the YAML file
- Exact formatting/indentation (will be reformatted)
- Field ordering (may change)

### Starting Services
After adding services via the API, you can start them using:
- **Individual service**: `POST /projects/{name}/services/{service}/start`
- **Docker Compose CLI**: `docker compose -p my-app up -d` in the project directory

The API uses Docker Compose CLI for orchestration, ensuring proper dependency handling, network creation, and volume management.
