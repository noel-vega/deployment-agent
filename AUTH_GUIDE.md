# Authentication Guide

This application uses JWT-based authentication with refresh token rotation, inspired by Azure's authentication model.

## Security Features

✅ **Short-lived access tokens** (5 minutes) - Limits damage window if compromised  
✅ **Long-lived refresh tokens** (7 days) - Better UX without frequent logins  
✅ **Token rotation** - New refresh token issued on each refresh (detects token reuse)  
✅ **httpOnly cookies** - Prevents XSS attacks from stealing tokens  
✅ **Secure cookies** - HTTPS-only in production  
✅ **SameSite protection** - Prevents CSRF attacks  
✅ **Bcrypt password hashing** - Industry-standard password security  
✅ **Server-side session tracking** - Revocable sessions (unlike pure stateless JWT)  

## Quick Start

### 1. Configuration

Copy the example environment file and configure:

```bash
cp .env.example .env
```

Edit `.env` and set strong secrets:

```bash
# Generate strong secrets (Linux/Mac)
openssl rand -base64 32

# Or use any random string generator
```

### 2. User Management

The admin user **must** be configured via environment variables. The application will not start without these credentials.

**Required environment variables:**

```bash
ADMIN_USERNAME=your-admin-username
ADMIN_PASSWORD=your-secure-password  # minimum 8 characters
```

⚠️ **REQUIRED**: You must set both `ADMIN_USERNAME` and `ADMIN_PASSWORD` before starting the application.

The password is automatically hashed with bcrypt at startup, so you can set plain-text passwords in the environment variables. The password must be at least 8 characters long.

### 3. Running the Application

```bash
# Development
export ADMIN_USERNAME=myadmin
export ADMIN_PASSWORD=mySecurePass123
go run main.go

# Or using .env file (recommended)
# 1. Copy .env.example to .env
# 2. Edit .env and set ADMIN_USERNAME and ADMIN_PASSWORD
# 3. Run: go run main.go

# Production
export ENVIRONMENT=production
export ADMIN_USERNAME=prodadmin
export ADMIN_PASSWORD=VerySecurePassword123!
./hubble
```

## API Endpoints

### Public Endpoints

#### Login
```bash
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response**: Sets `access_token` and `refresh_token` httpOnly cookies

```json
{
  "message": "Login successful",
  "expires_in": 300
}
```

#### Refresh Token
```bash
POST /auth/refresh
```

**Note**: Automatically uses `refresh_token` cookie. Returns new tokens.

```json
{
  "message": "Token refreshed successfully",
  "expires_in": 300
}
```

#### Logout
```bash
POST /auth/logout
```

Clears authentication cookies and revokes session.

```json
{
  "message": "Logged out successfully"
}
```

### Protected Endpoints

All protected endpoints require a valid `access_token` cookie.

#### List Containers
```bash
GET /containers
```

Returns list of Docker containers.

## Client Implementation

### Browser/Frontend Example

```javascript
// Login
async function login(username, password) {
  const response = await fetch('/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include', // Important: send cookies
    body: JSON.stringify({ username, password })
  });
  
  if (response.ok) {
    console.log('Logged in successfully');
    return true;
  }
  return false;
}

// Make authenticated requests
async function getContainers() {
  const response = await fetch('/containers', {
    credentials: 'include' // Important: send cookies
  });
  
  if (response.status === 401) {
    // Access token expired, try refresh
    await refreshToken();
    // Retry request
    return getContainers();
  }
  
  return response.json();
}

// Refresh token
async function refreshToken() {
  const response = await fetch('/auth/refresh', {
    method: 'POST',
    credentials: 'include'
  });
  
  if (response.status === 401) {
    // Refresh token expired, redirect to login
    window.location.href = '/login';
    return false;
  }
  
  return true;
}

// Logout
async function logout() {
  await fetch('/auth/logout', {
    method: 'POST',
    credentials: 'include'
  });
  window.location.href = '/login';
}
```

### cURL Examples

```bash
# Login
curl -X POST http://localhost:5000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# Access protected endpoint
curl http://localhost:5000/containers \
  -b cookies.txt

# Refresh token
curl -X POST http://localhost:5000/auth/refresh \
  -b cookies.txt \
  -c cookies.txt

# Logout
curl -X POST http://localhost:5000/auth/logout \
  -b cookies.txt
```

## Token Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User logs in                                             │
│    → Validates credentials                                  │
│    → Creates session                                        │
│    → Returns access_token (5 min) + refresh_token (7 days) │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. User makes API requests                                  │
│    → Sends access_token cookie automatically               │
│    → Server validates token                                │
│    → Returns protected data                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Access token expires (after 5 min)                       │
│    → Server returns 401 Unauthorized                        │
│    → Client calls /auth/refresh                            │
│    → Server validates refresh_token                        │
│    → Server issues NEW tokens (rotation)                   │
│    → Old refresh_token is revoked                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Refresh token expires (after 7 days)                     │
│    → Server returns 401 on /auth/refresh                   │
│    → Client redirects to login                             │
└─────────────────────────────────────────────────────────────┘
```

## Production Deployment

### Environment Variables

Set these in production:

```bash
export ENVIRONMENT=production
export ADMIN_USERNAME=your-admin-username
export ADMIN_PASSWORD=your-secure-password
export JWT_ACCESS_SECRET="$(openssl rand -base64 32)"
export JWT_REFRESH_SECRET="$(openssl rand -base64 32)"
export ACCESS_TOKEN_DURATION=5m
export REFRESH_TOKEN_DURATION=168h
```

### HTTPS Requirement

In production (`ENVIRONMENT=production`), cookies are set with the `Secure` flag, requiring HTTPS.

**Options**:
1. Use a reverse proxy (nginx, Caddy) with SSL/TLS
2. Use Let's Encrypt for free certificates
3. Deploy behind a load balancer with SSL termination

Example nginx config:

```nginx
server {
    listen 443 ssl;
    server_name dashboard.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Security Checklist

- [ ] Change default admin password
- [ ] Set strong JWT secrets (32+ characters)
- [ ] Enable HTTPS (set `ENVIRONMENT=production`)
- [ ] Store secrets in environment variables (never in code)
- [ ] Consider IP whitelisting if possible
- [ ] Monitor for suspicious login attempts
- [ ] Regularly rotate JWT secrets
- [ ] Keep dependencies updated

## Troubleshooting

### "Unauthorized - No access token"

- Ensure you're sending cookies with the request
- In browsers, use `credentials: 'include'`
- In cURL, use `-b cookies.txt`

### "Invalid credentials"

- Check username and password
- Verify `ADMIN_USERNAME` and `ADMIN_PASSWORD` in your `.env` file
- Ensure environment variables are loaded correctly

### Cookies not working

- Check `ENVIRONMENT` setting
- In production, ensure HTTPS is enabled
- Verify `SameSite` settings for your deployment

### Token expired immediately

- Check system clock synchronization
- Verify `ACCESS_TOKEN_DURATION` and `REFRESH_TOKEN_DURATION`

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│ Client (Browser/App)                                         │
│  - Stores tokens in httpOnly cookies                        │
│  - Automatically sends cookies with requests                │
│  - Handles 401 by refreshing tokens                         │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Middleware (middleware/auth.go)                              │
│  - Extracts access_token from cookie                        │
│  - Validates JWT signature and expiration                   │
│  - Adds user context to request                             │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Auth Service (auth/service.go)                               │
│  - Manages JWT token generation                             │
│  - Tracks active sessions in memory                         │
│  - Handles token rotation                                   │
│  - Cleans up expired sessions                               │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ User Store (auth/users.go)                                   │
│  - Loads admin user from environment variables              │
│  - Validates credentials against bcrypt hashes              │
│  - Hashes passwords automatically at startup                │
└──────────────────────────────────────────────────────────────┘
```

## FAQ

**Q: Why two tokens?**  
A: Short access tokens limit damage if stolen. Long refresh tokens provide better UX.

**Q: Why token rotation?**  
A: Detects token reuse (potential attack). If old refresh token is used, session is revoked.

**Q: Can I revoke sessions?**  
A: Yes, sessions are tracked server-side. Logout revokes the session.

**Q: How do I add database-backed users?**  
A: Replace `auth/users.go` with a database implementation. Keep bcrypt hashing.

**Q: Can I use this with a mobile app?**  
A: Yes, but you may want to store tokens differently (secure storage instead of cookies).

**Q: What about rate limiting?**  
A: Add go-chi's `middleware.Throttle` to `/auth/login` to prevent brute force.
