package routing

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hotel.com/bookingapi/internal/config"
	"hotel.com/bookingapi/internal/gateway/handlers"
	gwmiddleware "hotel.com/bookingapi/internal/gateway/middleware"
	"hotel.com/bookingapi/internal/logging"
)

// NewRouter creates and configures the chi router with all routes and middleware
func NewRouter(logger *logging.Logger, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Apply global middleware
	applyGlobalMiddleware(r, logger, cfg)

	// Create handlers
	h := handlers.NewHandler(logger, cfg)

	// Mount routes
	mountRoutes(r, h)

	return r
}

// applyGlobalMiddleware applies global middleware to the router
func applyGlobalMiddleware(r *chi.Mux, logger *logging.Logger, cfg *config.Config) {
	// Recovery middleware with panic recovery
	r.Use(middleware.Recoverer)

	// Request ID middleware
	r.Use(middleware.RequestID)

	// Logger middleware
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
func mountRoutes(r *chi.Mux, h *handlers.Handler) {
	// Health check routes (no auth required)
	r.Group(func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
		r.Get("/ready", h.ReadinessCheck)
		r.Get("/live", h.LivenessCheck)
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware to all API routes
		r.Use(gwmiddleware.NewAuthMiddleware().Middleware())

		// Booking routes
		r.Route("/bookings", func(r chi.Router) {
			r.Get("/", h.ListBookings)
			r.Post("/", h.CreateBooking)
			r.Get("/{id}", h.GetBooking)
			r.Put("/{id}", h.UpdateBooking)
			r.Delete("/{id}", h.DeleteBooking)
		})

		// Room routes
		r.Route("/rooms", func(r chi.Router) {
			r.Get("/", h.ListRooms)
			r.Get("/{id}", h.GetRoom)
			r.Post("/", h.CreateRoom)
			r.Put("/{id}", h.UpdateRoom)
			r.Delete("/{id}", h.DeleteRoom)
		})

		// Guest routes
		r.Route("/guests", func(r chi.Router) {
			r.Get("/", h.ListGuests)
			r.Get("/{id}", h.GetGuest)
			r.Post("/", h.CreateGuest)
			r.Put("/{id}", h.UpdateGuest)
			r.Delete("/{id}", h.DeleteGuest)
		})
	})

	// Proxy routes for upstream services
	r.Route("/proxy", func(r chi.Router) {
		r.Use(gwmiddleware.NewAuthMiddleware().Middleware())
		r.Route("/{service}", func(r chi.Router) {
			r.Handle("/*", h.ProxyHandler())
			r.Handle("/", h.ProxyHandler())
		})
	})
}
