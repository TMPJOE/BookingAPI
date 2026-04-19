package middleware

import (
	"net/http"
)

// CORS handles Cross-Origin Resource Sharing
type CORS struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// NewCORS creates a new CORS middleware with default settings
func NewCORS() *CORS {
	return &CORS{
		allowedOrigins:   []string{"*"},
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		allowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		exposedHeaders:   []string{"X-Request-Id"},
		allowCredentials: true,
		maxAge:           86400,
	}
}

// Middleware returns the CORS middleware
func (c *CORS) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				c.handlePreflight(w, r)
				return
			}

			// Set CORS headers for actual requests
			c.setHeaders(w, r)

			next.ServeHTTP(w, r)
		})
	}
}

func (c *CORS) handlePreflight(w http.ResponseWriter, r *http.Request) {
	c.setHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (c *CORS) setHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if c.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if c.allowedOrigins[0] == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	if c.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.Header().Set("Access-Control-Allow-Methods", joinStrings(c.allowedMethods))
	w.Header().Set("Access-Control-Allow-Headers", joinStrings(c.allowedHeaders))
	w.Header().Set("Access-Control-Expose-Headers", joinStrings(c.exposedHeaders))
	w.Header().Set("Access-Control-Max-Age", string(rune(c.maxAge)))
}

func (c *CORS) isOriginAllowed(origin string) bool {
	for _, allowed := range c.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
