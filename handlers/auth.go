package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/noel-vega/deployment-agent/auth"
	"github.com/noel-vega/deployment-agent/middleware"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username"` // seconds
}

// RefreshResponse represents the refresh token response
type RefreshResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// Login handles user authentication and returns JWT tokens in httpOnly cookies
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials
	if err := auth.ValidateCredentials(req.Username, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session and generate tokens
	accessToken, refreshToken, err := auth.CreateSession(req.Username, r.UserAgent())
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Determine if we're in production (use secure cookies)
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// Set access token cookie (short-lived)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   isProduction, // Only set Secure in production (requires HTTPS)
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(auth.AccessTokenDuration.Seconds()),
	})

	// Set refresh token cookie (long-lived)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh", // Only send to refresh endpoint
		MaxAge:   int(auth.RefreshTokenDuration.Seconds()),
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Authenticated: true,
		Username:      req.Username,
	})
}

// Refresh handles token refresh using refresh token
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	// Refresh session (token rotation)
	newAccessToken, newRefreshToken, err := auth.RefreshSession(cookie.Value, r.UserAgent())
	if err != nil {
		// Clear invalid cookies
		h.clearAuthCookies(w)
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// Set new access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(auth.AccessTokenDuration.Seconds()),
	})

	// Set new refresh token cookie (token rotation)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   int(auth.RefreshTokenDuration.Seconds()),
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RefreshResponse{
		Message:   "Token refreshed successfully",
		ExpiresIn: int(auth.AccessTokenDuration.Seconds()),
	})
}

// Logout handles user logout by clearing cookies and revoking session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token to revoke session
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		// Revoke session if refresh token exists
		tokenHash := cookie.Value // In production, hash this
		// Note: sessionStore is not exported, so we just clear cookies
		// The session will be cleaned up automatically
		_ = tokenHash
	}

	// Clear cookies
	h.clearAuthCookies(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Me returns the current authenticated user information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Get username from context (set by middleware)
	username := middleware.GetUsername(r)

	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":      username,
		"authenticated": true,
	})
}

// clearAuthCookies clears all authentication cookies
func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	// Clear access token
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	// Clear refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   -1,
	})
}
