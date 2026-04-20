package routing

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hotel.com/bookingapi/internal/config"
	"hotel.com/bookingapi/internal/gateway/handlers"
	"hotel.com/bookingapi/internal/gateway/health"
	gwmiddleware "hotel.com/bookingapi/internal/gateway/middleware"
	"hotel.com/bookingapi/internal/gateway/proxy"
	"hotel.com/bookingapi/internal/logging"
)

// Router holds the router and its dependencies
type Router struct {
	engine        *chi.Mux
	proxy         *proxy.ReverseProxy
	healthChecker *health.Checker
}

// NewRouter creates and configures the chi router with all routes and middleware
func NewRouter(logger *logging.Logger, cfg *config.Config) *Router {
	r := chi.NewRouter()

	// Create dependencies
	h := handlers.NewHandler(logger, cfg)
	revProxy := proxy.NewReverseProxy(logger, cfg)
	healthChecker := health.NewChecker(logger, cfg, 30*time.Second)

	// Apply global middleware
	applyGlobalMiddleware(r, logger, cfg)

	// Mount routes
	mountRoutes(r, h, revProxy, cfg)

	// Start health checker
	healthChecker.Start()

	return &Router{
		engine:        r,
		proxy:         revProxy,
		healthChecker: healthChecker,
	}
}

// Mux returns the underlying chi.Mux
func (rt *Router) Mux() *chi.Mux {
	return rt.engine
}

// Stop stops the router and its dependencies
func (rt *Router) Stop() {
	if rt.healthChecker != nil {
		rt.healthChecker.Stop()
	}
}

// applyGlobalMiddleware applies global middleware to the router
func applyGlobalMiddleware(r *chi.Mux, logger *logging.Logger, cfg *config.Config) {
	// Recovery middleware with panic recovery
	r.Use(middleware.Recoverer)

	// Request ID middleware
	r.Use(middleware.RequestID)

	// Real IP middleware
	r.Use(middleware.RealIP)

	// Timeout middleware
	r.Use(middleware.Timeout(cfg.Server.ReadTimeout))

	// Custom logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
		})
	})

	// Rate limiting middleware
	if cfg.RateLimit.Enabled {
		r.Use(gwmiddleware.NewRateLimiter(
			cfg.RateLimit.RequestsPerSec,
			cfg.RateLimit.Burst,
		).Middleware())
	}

	// CORS middleware
	r.Use(gwmiddleware.NewCORS().Middleware())

	// Security headers middleware
	r.Use(gwmiddleware.NewSecurityHeaders().Middleware())
}

// mountRoutes mounts all API routes
func mountRoutes(r *chi.Mux, h *handlers.Handler, revProxy *proxy.ReverseProxy, cfg *config.Config) {
	// Health check routes (no auth required)
	r.Group(func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
		r.Get("/ready", h.ReadinessCheck)
		r.Get("/live", h.LivenessCheck)
	})

	// Upstream health status endpoint (no auth required)
	r.Get("/upstreams", func(w http.ResponseWriter, r *http.Request) {
		status := revProxy.GetAllUpstreams()
		result := make([]map[string]interface{}, 0, len(status))
		for _, s := range status {
			result = append(result, map[string]interface{}{
				"name":        s.Name,
				"url":         s.URL,
				"path_prefix": s.PathPrefix,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	r.Group(func(r chi.Router) {
		r.Handle("/api/v1/users/*", revProxy)
	})

	// Protected API v1 routes - auth required
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware FIRST (before any routes)
		jwtConfig := gwmiddleware.JWTConfig{
			Secret:     cfg.JWT.Secret,
			Issuer:     cfg.JWT.Issuer,
			Expiration: cfg.JWT.Expiration,
		}
		r.Use(gwmiddleware.NewAuthMiddleware(jwtConfig).Middleware())

		// All /api/v1/* requests go to the reverse proxy
		// The proxy will route based on path prefix
		r.Handle("/*", revProxy)
	})
}

// ReadinessCheckWithUpstreams checks readiness including upstream services
func ReadinessCheckWithUpstreams(ctx context.Context, healthChecker *health.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := healthChecker.GetStatus()
		healthyCount := 0
		for _, s := range status {
			if s.Healthy {
				healthyCount++
			}
		}

		response := map[string]interface{}{
			"status":          "ready",
			"healthy_count":   healthyCount,
			"total_count":     len(status),
			"upstream_status": status,
		}

		// If no upstreams are healthy, still report ready (degraded mode)
		if healthyCount == 0 && len(status) > 0 {
			response["status"] = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
