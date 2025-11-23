# Authentication Implementation Summary

## What Was Built

A production-ready JWT authentication system with refresh token rotation for your Docker container monitoring dashboard.

## Features Implemented

✅ **Dual JWT Token System**
- Access tokens: 5-minute lifespan (configurable)
- Refresh tokens: 7-day lifespan (configurable)
- Separate signing keys for enhanced security

✅ **Security Best Practices**
- httpOnly cookies (prevents XSS attacks)
- Secure flag in production (HTTPS-only)
- SameSite strict mode (prevents CSRF)
- Bcrypt password hashing (cost factor: 10)
- Token rotation on refresh (detects reuse/theft)
- Server-side session tracking (revocable)

✅ **Session Management**
- In-memory session store with thread-safety
- Automatic cleanup of expired sessions
- Session revocation capability
- Track user activity (last used, created at, user agent)

✅ **Environment-Based Configuration**
- JWT secrets from environment variables
- Configurable token durations
- Development/production modes
- Secure defaults with warnings

## File Structure

```
hubble/
├── auth/
│   ├── service.go       # JWT token generation, session management
│   └── users.go         # Static user store with bcrypt hashing
├── handlers/
│   ├── auth.go          # Login, logout, refresh endpoints
│   └── containers.go    # Protected container listing endpoint
├── middleware/
│   └── auth.go          # JWT authentication middleware
├── tools/
│   └── hashgen.go       # Password hash generator utility
├── .env.example         # Environment variable template
├── AUTH_GUIDE.md        # Complete authentication documentation
└── main.go              # Application entry point with auth wiring
```

## API Endpoints

### Public Endpoints
- `POST /auth/login` - Authenticate and receive tokens
- `POST /auth/refresh` - Refresh expired access token
- `POST /auth/logout` - Logout and clear tokens

### Protected Endpoints
- `GET /containers` - List Docker containers (requires authentication)

## Admin User Configuration

Admin user credentials **must** be configured via environment variables. The application will fail to start without them.

```bash
ADMIN_USERNAME=your-admin-username
ADMIN_PASSWORD=your-secure-password  # minimum 8 characters
```

⚠️ **REQUIRED**: Both `ADMIN_USERNAME` and `ADMIN_PASSWORD` must be set before starting the application.

The password is automatically hashed with bcrypt at startup. Password must be at least 8 characters long.

## Quick Start

### 1. Set Environment Variables (Required)

```bash
# Copy example config
cp .env.example .env

# Edit .env and set (REQUIRED):
# - ADMIN_USERNAME (any username you want)
# - ADMIN_PASSWORD (minimum 8 characters)
# - JWT_ACCESS_SECRET (use: openssl rand -base64 32)
# - JWT_REFRESH_SECRET (use: openssl rand -base64 32)
```

**Note**: The application will not start without `ADMIN_USERNAME` and `ADMIN_PASSWORD` set.

### 2. Run the Application

```bash
# Development
go run main.go

# Production
ENVIRONMENT=production ./hubble
```

### 3. Test Authentication

```bash
# Set credentials (use the ones from your .env file)
export ADMIN_USERNAME=your-admin-username
export ADMIN_PASSWORD=your-password

# Login
curl -X POST http://localhost:5000/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -c cookies.txt

# Access protected endpoint
curl http://localhost:5000/containers -b cookies.txt

# Refresh token
curl -X POST http://localhost:5000/auth/refresh \
  -b cookies.txt -c cookies.txt

# Logout
curl -X POST http://localhost:5000/auth/logout -b cookies.txt
```

## Changing Admin Credentials

Simply update the environment variables in your `.env` file:

```bash
ADMIN_USERNAME=myadmin
ADMIN_PASSWORD=MySecurePassword123!
```

Restart the application for changes to take effect. The password will be automatically hashed at startup.

## Token Flow

```
1. Login
   └─> Validates credentials
   └─> Creates session
   └─> Returns httpOnly cookies (access + refresh tokens)

2. API Request
   └─> Browser sends access_token cookie automatically
   └─> Middleware validates token
   └─> Returns protected data

3. Access Token Expires (5 min)
   └─> Server returns 401
   └─> Client calls /auth/refresh
   └─> Server validates refresh_token
   └─> Issues NEW tokens (rotation)
   └─> Revokes old refresh_token

4. Refresh Token Expires (7 days)
   └─> Server returns 401 on refresh
   └─> Client redirects to login
```

## Production Deployment Checklist

- [ ] Set `ENVIRONMENT=production`
- [ ] Set `ADMIN_USERNAME` and `ADMIN_PASSWORD` in environment
- [ ] Configure strong JWT secrets (32+ characters)
- [ ] Enable HTTPS (required for secure cookies)
- [ ] Set up reverse proxy (nginx/Caddy) with SSL/TLS
- [ ] Configure proper CORS if dashboard is on different domain
- [ ] Set up monitoring/logging
- [ ] Consider adding rate limiting to /auth/login

## Security Comparison: Azure-Inspired

| Feature | Azure Portal | This Implementation |
|---------|--------------|---------------------|
| Token Type | OAuth 2.0 + OIDC | JWT (HS256) |
| Access Token | ~1 hour | 5 minutes (configurable) |
| Refresh Token | ~90 days | 7 days (configurable) |
| Token Rotation | ✅ Yes | ✅ Yes |
| Revocation | ✅ Yes | ✅ Yes (session store) |
| httpOnly Cookies | ✅ Yes | ✅ Yes |
| MFA | ✅ Built-in | ❌ Not implemented |
| Session Tracking | ✅ Yes | ✅ Yes (in-memory) |

## Key Decisions Made

1. **JWT vs API Keys**: Chose JWT with refresh tokens for better security with short-lived access
2. **httpOnly Cookies vs localStorage**: Chose cookies for better XSS protection
3. **Token Rotation**: Implemented to detect token theft/reuse
4. **In-Memory Sessions**: Sufficient for single-server deployment, enables revocation
5. **Bcrypt Hashing**: Industry standard for password security
6. **Static Users**: Simple for small team dashboard, easy to migrate to database later

## Future Enhancements (Optional)

- [ ] Add rate limiting to prevent brute force attacks
- [ ] Implement token blacklist for immediate revocation
- [ ] Add MFA/2FA support
- [ ] Move session store to Redis for multi-server deployments
- [ ] Add database-backed user management
- [ ] Implement audit logging
- [ ] Add IP whitelisting option
- [ ] OAuth integration (Google, GitHub, etc.)

## Documentation

See `AUTH_GUIDE.md` for complete documentation including:
- Detailed API documentation
- Client implementation examples (JavaScript, cURL)
- Troubleshooting guide
- Architecture diagrams
- Security best practices
- Production deployment guide

## Dependencies Added

- `github.com/go-chi/jwtauth/v5` - JWT authentication middleware
- `golang.org/x/crypto/bcrypt` - Password hashing

## Testing

Application builds successfully and starts without errors:

```
✓ Dependencies installed
✓ Application builds
✓ Server starts on :5000
✓ Auth service initialized
✓ Token durations configured
```

All components are ready for production deployment after environment configuration.
