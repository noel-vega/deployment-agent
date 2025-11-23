package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/noel-vega/deployment-agent/auth"
)

// Protected is a middleware that validates JWT access tokens from cookies
func Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get access token from cookie
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized - No access token", http.StatusUnauthorized)
			return
		}

		// Verify and decode the token
		token, err := jwtauth.VerifyToken(auth.AccessTokenAuth, cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		// Check if token is valid
		if token == nil {
			http.Error(w, "Unauthorized - Token validation failed", http.StatusUnauthorized)
			return
		}

		// Add token to context for downstream handlers
		ctx := jwtauth.NewContext(r.Context(), token, nil)

		// Extract username from token and add to context
		if username, ok := token.Get("username"); ok {
			ctx = context.WithValue(ctx, "username", username)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUsername extracts username from the request context
func GetUsername(r *http.Request) string {
	username, ok := r.Context().Value("username").(string)
	if !ok {
		return ""
	}
	return username
}
