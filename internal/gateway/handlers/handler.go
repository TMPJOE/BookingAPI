package handlers

import (
	"encoding/json"
	"net/http"

	"hotel.com/bookingapi/internal/config"
	"hotel.com/bookingapi/internal/logging"
)

// Handler holds all HTTP handlers for the thin gateway
type Handler struct {
	logger *logging.Logger
	config *config.Config
}

// NewHandler creates a new handler instance
func NewHandler(logger *logging.Logger, cfg *config.Config) *Handler {
	return &Handler{
		logger: logger,
		config: cfg,
	}
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "api-gateway",
		"version": "1.0.0",
	}
	respondJSON(w, http.StatusOK, response)
}

// ReadinessCheck handles GET /ready
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Add actual readiness checks (upstream services connectivity)
	response := map[string]string{
		"status": "ready",
	}
	respondJSON(w, http.StatusOK, response)
}

// LivenessCheck handles GET /live
func (h *Handler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "alive",
	}
	respondJSON(w, http.StatusOK, response)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
