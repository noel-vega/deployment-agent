# Development Guide

Guide for developers working on Hubble.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Building](#building)
- [Contributing](#contributing)

## Development Setup

### Prerequisites

- **Go 1.24+**
- **Docker 20.10+**
- **Docker Compose v2+**
- **Make** (optional, for convenience commands)
- **Git**

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/noel-vega/hubble
cd hubble

# Install dependencies
go mod download

# Install development tools
go install github.com/air-verse/air@latest  # Hot reload

# Copy environment template
cp .env.example .env

# Edit .env with development credentials
nano .env
```

### Development Environment

```bash
# .env for development
ENVIRONMENT=development
ADMIN_USERNAME=admin
ADMIN_PASSWORD=devpass123
JWT_ACCESS_SECRET=dev-secret-change-in-production
JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
ACCESS_TOKEN_DURATION=5m
REFRESH_TOKEN_DURATION=168h
PROJECTS_ROOT_PATH=./projects
HUBBLE_TRAEFIK_ENABLED=false  # Enable if testing Traefik
```

### Start Development Server

```bash
# Option 1: Using Make (with hot reload)
make dev

# Option 2: Using Air directly
air

# Option 3: Manual
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=devpass123
go run main.go
```

Server will start on `http://localhost:5000`

## Project Structure

```
hubble/
├── .github/
│   └── workflows/          # CI/CD workflows
├── auth/
│   ├── service.go          # JWT token management
│   └── users.go            # User authentication
├── docker/
│   └── service.go          # Docker client wrapper
├── handlers/
│   ├── auth.go             # Authentication endpoints
│   ├── containers.go       # Container management endpoints
│   ├── images.go           # Image listing endpoints
│   ├── projects.go         # Project management endpoints
│   └── registry.go         # Registry browsing endpoints
├── middleware/
│   └── auth.go             # Authentication middleware
├── platform/
│   ├── infrastructure.go   # Network auto-creation
│   └── traefik.go          # Traefik auto-provisioning
├── projects/
│   └── service.go          # Docker Compose project management
├── registry/
│   └── client.go           # Docker registry client
├── main.go                 # Application entry point
├── Dockerfile              # Production Docker image
├── docker-compose.yml      # Development/deployment compose
├── .air.toml               # Hot reload configuration
├── Makefile                # Development commands
├── go.mod                  # Go dependencies
└── docs/
    ├── README.md           # Main documentation
    ├── SETUP.md            # Setup guide
    ├── API.md              # API reference
    ├── TRAEFIK.md          # Traefik integration
    └── DEVELOPMENT.md      # This file
```

### Key Files

**main.go**
- Application entry point
- Initializes services
- Sets up routes
- Calls platform infrastructure setup

**platform/infrastructure.go**
- Auto-creates `hubble` network
- Coordinates platform setup

**platform/traefik.go**
- Auto-provisions Traefik container
- Manages Traefik lifecycle

**auth/service.go**
- JWT token generation
- Session management
- Token rotation

**projects/service.go**
- Docker Compose file manipulation
- Service/network CRUD operations
- Container orchestration

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/my-feature
```

### 2. Make Changes

Edit code with hot reload enabled:

```bash
# Terminal 1: Run dev server
make dev

# Terminal 2: Test your changes
curl http://localhost:5000/your-endpoint
```

### 3. Test Manually

```bash
# Login
curl -X POST http://localhost:5000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"devpass123"}' \
  -c cookies.txt

# Test your endpoint
curl http://localhost:5000/your-endpoint \
  -b cookies.txt
```

### 4. Build and Test

```bash
# Build
make build

# Run tests (when available)
make test

# Test the binary
./hubble-server
```

### 5. Commit and Push

```bash
git add .
git commit -m "feat: add new feature"
git push origin feature/my-feature
```

## Testing

### Manual Testing

```bash
# Start server
make dev

# In another terminal, test authentication
make test-auth

# Test specific endpoints
curl http://localhost:5000/projects \
  -b cookies.txt
```

### Testing Traefik Integration

```bash
# Enable Traefik in .env
HUBBLE_TRAEFIK_ENABLED=true
HUBBLE_TRAEFIK_EMAIL=test@example.com

# Restart
make dev

# Check Traefik is running
docker ps | grep traefik

# Check logs
docker logs hubble-traefik
```

### Testing Docker Integration

```bash
# List containers
curl http://localhost:5000/containers -b cookies.txt

# Create test project
curl -X POST http://localhost:5000/projects \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"test"}'

# Add service
curl -X POST http://localhost:5000/projects/test/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "nginx",
    "image": "nginx:alpine",
    "ports": ["8080:80"]
  }'

# Start service
curl -X POST http://localhost:5000/projects/test/services/nginx/start \
  -b cookies.txt

# Verify
curl http://localhost:8080
```

## Building

### Build Binary

```bash
# Using Make
make build

# Manual
go build -o hubble-server .

# With optimizations
go build -ldflags="-s -w" -o hubble-server .
```

### Build Docker Image

```bash
# Using Make
make docker-build

# Manual
docker build -t hubble-server .

# With custom tag
docker build -t myregistry.com/hubble-server:v1.0 .
```

### Build for Production

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hubble-server-linux-amd64 .

# Linux ARM64 (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o hubble-server-linux-arm64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o hubble-server-darwin-amd64 .
```

## Available Make Commands

```bash
make help          # Show all available commands
make dev           # Run development server with hot reload
make build         # Build the application
make run           # Build and run
make test          # Run tests
make test-auth     # Test authentication flow
make clean         # Clean build artifacts
make deps          # Install dependencies
make hash          # Generate bcrypt password hash
make docker-build  # Build Docker image
make docker-run    # Run with docker-compose
make docker-stop   # Stop Docker containers
make fmt           # Format code
make lint          # Run linter (if configured)
```

## Code Style

### Go Formatting

```bash
# Format all code
go fmt ./...

# Or using Make
make fmt
```

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `auth`, `platform`)
- **Files**: lowercase with underscores (e.g., `auth_service.go`)
- **Exported functions**: PascalCase (e.g., `CreateProject`)
- **Unexported functions**: camelCase (e.g., `ensureNetwork`)
- **Constants**: PascalCase (e.g., `HubbleNetworkName`)

### Error Handling

Always return errors, don't panic:

```go
// Good
func DoSomething() error {
    if err := operation(); err != nil {
        return fmt.Errorf("failed to do something: %w", err)
    }
    return nil
}

// Bad
func DoSomething() {
    if err := operation(); err != nil {
        panic(err)  // Don't do this
    }
}
```

### Logging

Use `log.Printf` for important events:

```go
log.Printf("✓ Created Hubble network '%s' (ID: %s)", name, id[:12])
log.Printf("Warning: Traefik email not set")
log.Printf("Failed to create project: %v", err)
```

## Contributing

### Before Submitting a PR

1. **Format your code**
   ```bash
   go fmt ./...
   ```

2. **Build successfully**
   ```bash
   make build
   ```

3. **Test manually**
   - Test all affected endpoints
   - Verify no regressions

4. **Update documentation**
   - Update API.md if adding/changing endpoints
   - Update relevant .md files

5. **Write clear commit messages**
   ```
   feat: add project deletion endpoint
   fix: correct network validation logic
   docs: update Traefik configuration examples
   refactor: simplify auth middleware
   ```

### Commit Message Format

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Changes to build process or auxiliary tools

**Examples:**

```
feat: add service update endpoint

Allows updating service configuration via PUT request.
Validates all fields and preserves unspecified fields.

Closes #123
```

```
fix: correct port binding validation

Previously allowed invalid port formats like "abc:80".
Now properly validates port numbers and ranges.
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit PR with clear description
6. Address review feedback
7. Squash commits if requested

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tested manually
- [ ] Added/updated tests
- [ ] All tests passing

## Checklist
- [ ] Code formatted with `go fmt`
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

## Debugging

### Enable Verbose Logging

```bash
# Set log level
export LOG_LEVEL=debug

# Run with verbose output
go run main.go
```

### Debug Docker Issues

```bash
# Check Docker is running
docker info

# Check Docker socket
ls -l /var/run/docker.sock

# Test Docker client
docker ps

# View container logs
docker logs hubble-server
docker logs hubble-traefik
```

### Debug Authentication Issues

```bash
# Generate password hash manually
make hash PASSWORD=mypassword

# Test login with verbose output
curl -v -X POST http://localhost:5000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"devpass123"}' \
  -c cookies.txt

# Check cookies
cat cookies.txt
```

### Debug Traefik Issues

```bash
# Check Traefik logs
docker logs hubble-traefik

# Check Traefik configuration
docker exec hubble-traefik cat /etc/traefik/traefik.yml

# Check network connectivity
docker network inspect hubble

# Access dashboard
# SSH tunnel if needed:
ssh -L 8080:localhost:8080 user@your-server
# Then visit: http://localhost:8080
```

## IDE Setup

### VS Code

Recommended extensions:
- **Go** (golang.go)
- **Docker** (ms-azuretools.vscode-docker)
- **YAML** (redhat.vscode-yaml)

Settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofmt",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### GoLand / IntelliJ IDEA

1. Open project
2. Enable Go modules support
3. Configure run configuration:
   - Program arguments: (none)
   - Environment: Load from `.env`
4. Enable format on save

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Docker SDK for Go](https://docs.docker.com/engine/api/sdk/)
- [Chi Router](https://github.com/go-chi/chi)
- [JWT Auth](https://github.com/go-chi/jwtauth)
- [Traefik Documentation](https://doc.traefik.io/traefik/)

## Getting Help

- **Issues**: https://github.com/noel-vega/hubble/issues
- **Discussions**: https://github.com/noel-vega/hubble/discussions
- **Email**: noel@example.com (if applicable)

## License

MIT License - see LICENSE file for details
