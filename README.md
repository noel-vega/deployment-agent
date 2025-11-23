# Hubble - Docker Container Dashboard

A secure, production-ready API for monitoring Docker containers with JWT authentication.

## Quick Start

### Development (Easiest)

```bash
# Start development server with hot reload
make dev

# Server runs at http://localhost:5000
# Default credentials: admin / devpass123
```

The `make dev` command automatically sets up development environment variables and runs with hot reload using air.

### Production Setup

#### 1. Set Required Environment Variables

```bash
cp .env.example .env
```

Edit `.env` and configure:

```bash
# REQUIRED - Application will not start without these
ADMIN_USERNAME=your-admin-username
ADMIN_PASSWORD=your-secure-password  # minimum 8 characters

# REQUIRED for production - Generate strong random secrets
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)

# Optional - Configure token durations
ACCESS_TOKEN_DURATION=5m
REFRESH_TOKEN_DURATION=168h

# Set to production for HTTPS-only cookies
ENVIRONMENT=production
```

#### 2. Run the Application

```bash
# Build and run
make build
make run

# Or manually
go build -o hubble .
ENVIRONMENT=production ./hubble
```

### 3. Authentication

Login to get JWT tokens:

```bash
curl -X POST http://localhost:5000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your-admin-username","password":"your-password"}' \
  -c cookies.txt
```

Access protected endpoints:

```bash
curl http://localhost:5000/containers -b cookies.txt
```

## Features

✅ **Secure JWT Authentication** with short-lived access tokens (5 min) and refresh token rotation  
✅ **httpOnly Cookies** - XSS protection  
✅ **Bcrypt Password Hashing** - Industry-standard security  
✅ **Server-side Session Tracking** - Revocable sessions  
✅ **Environment-based Configuration** - No hardcoded secrets  
✅ **Production-ready** - HTTPS support, secure cookies  

## API Endpoints

### Authentication
- `POST /auth/login` - Login and receive tokens
- `POST /auth/refresh` - Refresh expired access token
- `POST /auth/logout` - Logout and clear tokens

### Protected Endpoints
- `GET /containers` - List Docker containers (requires authentication)

## Documentation

- **[AUTH_GUIDE.md](AUTH_GUIDE.md)** - Complete authentication documentation
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - Technical implementation details
- **[.env.example](.env.example)** - Environment configuration template

## Security

This application implements Azure-inspired authentication with:
- Short-lived access tokens (5 minutes)
- Long-lived refresh tokens (7 days)
- Token rotation on refresh
- Server-side session management
- Minimum 8-character password requirement
- No default credentials (fail-safe)

## Requirements

- Go 1.24+
- Docker (for container monitoring)
- Environment variables for admin credentials

## Available Make Commands

```bash
make help          # Show all available commands
make dev           # Run development server with hot reload
make build         # Build the application
make run           # Build and run
make test          # Run tests
make test-auth     # Test authentication flow (requires running server)
make clean         # Clean build artifacts
make deps          # Install dependencies
make hash          # Generate bcrypt password hash
make docker-build  # Build Docker image
make docker-run    # Run with docker-compose
make docker-stop   # Stop Docker container
```

## Development Workflow

```bash
# 1. Start dev server with hot reload
make dev

# 2. Make changes to code - server auto-restarts

# 3. Test authentication in another terminal
make test-auth

# 4. Generate password hashes if needed
make hash PASSWORD=mynewpassword
```

## Testing

```bash
# With make (recommended)
make dev  # Start server in one terminal
make test-auth  # Run auth tests in another terminal

# Manual testing
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=devpass123
./test_auth.sh
```

## License

MIT
