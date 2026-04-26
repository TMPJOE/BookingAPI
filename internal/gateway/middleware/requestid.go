package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}

// requestIDKey is the context key for request ID
const requestIDKey = "X-Request-Id"

// GenerateRequestID generates a new UUID v7 for request identification
func GenerateRequestID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to random UUID if V7 fails
		id = uuid.New()
	}
	return id.String()
}
