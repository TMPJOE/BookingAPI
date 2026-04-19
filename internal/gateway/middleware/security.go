package middleware

import (
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers
type SecurityHeaders struct{}

// NewSecurityHeaders creates a new security headers middleware
func NewSecurityHeaders() *SecurityHeaders {
	return &SecurityHeaders{}
}

// Middleware returns the security headers middleware
func (s *SecurityHeaders) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict Transport Security (HSTS)
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// Content Security Policy
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}
