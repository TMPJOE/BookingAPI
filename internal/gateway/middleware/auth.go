package middleware

import (
	"net/http"
	"strings"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	apiKeys map[string]bool // In production, use a proper auth service
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		apiKeys: make(map[string]bool),
	}
}

// Middleware returns the authentication middleware
func (a *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check endpoints
			if isHealthEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			if !a.validateToken(token) {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validateToken validates the authentication token
// In production, this should verify JWT or call an auth service
func (a *AuthMiddleware) validateToken(token string) bool {
	// Placeholder: In production, implement proper token validation
	// For now, accept any non-empty token
	return len(token) > 0
}

// isHealthEndpoint checks if the path is a health check endpoint
func isHealthEndpoint(path string) bool {
	healthPaths := []string{"/health", "/ready", "/live"}
	for _, p := range healthPaths {
		if path == p || strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
