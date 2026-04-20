package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context keys for JWT claims
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	ClaimsKey    contextKey = "claims"
)

// JWT configuration errors
var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
	ErrMissingToken  = errors.New("missing token")
	ErrInvalidFormat = errors.New("invalid authorization format")
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Issuer     string
	Expiration time.Duration
}

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	config JWTConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
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
				respondError(w, http.StatusUnauthorized, ErrMissingToken.Error())
				return
			}

			// Check Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondError(w, http.StatusUnauthorized, ErrInvalidFormat.Error())
				return
			}

			tokenString := parts[1]
			claims, err := a.ValidateToken(tokenString)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					respondError(w, http.StatusUnauthorized, ErrTokenExpired.Error())
					return
				}
				respondError(w, http.StatusUnauthorized, ErrInvalidToken.Error())
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthMiddleware) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(a.config.Secret), nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer if configured
	if a.config.Issuer != "" {
		issuer, err := claims.GetIssuer()
		if err != nil || issuer != a.config.Issuer {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a user
func (a *AuthMiddleware) GenerateToken(userID, email string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(a.config.Expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.Secret))
}

// GetUserIDFromContext extracts user ID from the request context
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserEmailFromContext extracts user email from the request context
func GetUserEmailFromContext(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

// GetClaimsFromContext extracts JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) *JWTClaims {
	if claims, ok := ctx.Value(ClaimsKey).(*JWTClaims); ok {
		return claims
	}
	return nil
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

// ErrorResponse is the standard JSON shape for all error responses
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError writes a consistent JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    http.StatusText(status),
			Message: message,
		},
	})
}

// HashPassword creates a SHA256 hash of the password (for demonstration)
// In production, use bcrypt or argon2
func HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}
