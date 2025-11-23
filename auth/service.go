package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

var (
	// AccessTokenAuth handles short-lived access tokens (5 minutes)
	AccessTokenAuth *jwtauth.JWTAuth
	// RefreshTokenAuth handles long-lived refresh tokens (7 days)
	RefreshTokenAuth *jwtauth.JWTAuth

	// Token duration configurations
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
)

// Session represents an active user session
type Session struct {
	Username         string
	RefreshTokenHash string
	CreatedAt        time.Time
	LastUsedAt       time.Time
	UserAgent        string
}

// SessionStore manages active sessions in memory
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session // key: refresh token hash
}

var sessionStore = &SessionStore{
	sessions: make(map[string]*Session),
}

// Initialize sets up JWT auth instances and loads configuration
func Initialize() error {
	// Load secrets from environment
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		accessSecret = "default-access-secret-change-this-in-production"
		fmt.Println("WARNING: Using default JWT_ACCESS_SECRET. Set JWT_ACCESS_SECRET environment variable in production!")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = "default-refresh-secret-change-this-in-production"
		fmt.Println("WARNING: Using default JWT_REFRESH_SECRET. Set JWT_REFRESH_SECRET environment variable in production!")
	}

	// Initialize JWT auth instances
	AccessTokenAuth = jwtauth.New("HS256", []byte(accessSecret), nil)
	RefreshTokenAuth = jwtauth.New("HS256", []byte(refreshSecret), nil)

	// Load token durations
	accessDuration := os.Getenv("ACCESS_TOKEN_DURATION")
	if accessDuration == "" {
		accessDuration = "5m"
	}
	var err error
	AccessTokenDuration, err = time.ParseDuration(accessDuration)
	if err != nil {
		return fmt.Errorf("invalid ACCESS_TOKEN_DURATION: %w", err)
	}

	refreshDuration := os.Getenv("REFRESH_TOKEN_DURATION")
	if refreshDuration == "" {
		refreshDuration = "168h" // 7 days
	}
	RefreshTokenDuration, err = time.ParseDuration(refreshDuration)
	if err != nil {
		return fmt.Errorf("invalid REFRESH_TOKEN_DURATION: %w", err)
	}

	// Start session cleanup goroutine
	go sessionStore.cleanupExpiredSessions()

	return nil
}

// GenerateAccessToken creates a short-lived access token
func GenerateAccessToken(username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(AccessTokenDuration)

	claims := map[string]interface{}{
		"username": username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	_, tokenString, err := AccessTokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a long-lived refresh token
func GenerateRefreshToken(username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(RefreshTokenDuration)

	// Generate unique token ID for tracking
	tokenID, err := generateRandomString(32)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token ID: %w", err)
	}

	claims := map[string]interface{}{
		"username": username,
		"token_id": tokenID,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	_, tokenString, err := RefreshTokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// CreateSession creates a new session and returns both tokens
func CreateSession(username, userAgent string) (accessToken, refreshToken string, err error) {
	// Generate tokens
	accessToken, _, err = GenerateAccessToken(username)
	if err != nil {
		return "", "", err
	}

	refreshToken, _, err = GenerateRefreshToken(username)
	if err != nil {
		return "", "", err
	}

	// Store session
	session := &Session{
		Username:         username,
		RefreshTokenHash: hashToken(refreshToken),
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
		UserAgent:        userAgent,
	}

	sessionStore.mu.Lock()
	sessionStore.sessions[session.RefreshTokenHash] = session
	sessionStore.mu.Unlock()

	return accessToken, refreshToken, nil
}

// RefreshSession validates refresh token and issues new tokens (token rotation)
func RefreshSession(oldRefreshToken, userAgent string) (newAccessToken, newRefreshToken string, err error) {
	tokenHash := hashToken(oldRefreshToken)

	// Verify session exists
	sessionStore.mu.RLock()
	_, exists := sessionStore.sessions[tokenHash]
	sessionStore.mu.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("invalid or expired refresh token")
	}

	// Verify JWT is valid
	token, err := RefreshTokenAuth.Decode(oldRefreshToken)
	if err != nil || token == nil {
		// Token is invalid, remove session
		sessionStore.RevokeSession(tokenHash)
		return "", "", fmt.Errorf("invalid refresh token")
	}

	// Extract username from token
	username, ok := token.Get("username")
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}

	// Revoke old refresh token (rotation)
	sessionStore.RevokeSession(tokenHash)

	// Create new session with new tokens
	newAccessToken, newRefreshToken, err = CreateSession(username.(string), userAgent)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// RevokeSession removes a session from the store
func (s *SessionStore) RevokeSession(tokenHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, tokenHash)
}

// RevokeAllUserSessions removes all sessions for a specific user
func (s *SessionStore) RevokeAllUserSessions(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hash, session := range s.sessions {
		if session.Username == username {
			delete(s.sessions, hash)
		}
	}
}

// cleanupExpiredSessions periodically removes expired sessions
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for hash, session := range s.sessions {
			// Remove sessions older than refresh token duration
			if now.Sub(session.CreatedAt) > RefreshTokenDuration {
				delete(s.sessions, hash)
			}
		}
		s.mu.Unlock()
	}
}

// GetSessionCount returns the number of active sessions
func (s *SessionStore) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// hashToken creates a hash of the token for storage (simple hash for lookup)
func hashToken(token string) string {
	// For simplicity, using the token itself as key
	// In production, consider using SHA256
	return token
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
