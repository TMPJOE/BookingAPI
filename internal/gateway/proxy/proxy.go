package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"hotel.com/bookingapi/internal/config"
	"hotel.com/bookingapi/internal/logging"
)

// ReverseProxy handles routing requests to upstream services
type ReverseProxy struct {
	logger    *logging.Logger
	config    *config.Config
	upstreams map[string]*url.URL // path prefix -> upstream URL
	director  func(*http.Request)
}

// NewReverseProxy creates a new reverse proxy instance
func NewReverseProxy(logger *logging.Logger, cfg *config.Config) *ReverseProxy {
	upstreams := make(map[string]*url.URL)

	// Build upstream map from config
	for _, upstream := range cfg.Upstreams {
		if upstream.URL != "" && upstream.PathPrefix != "" {
			parsedURL, err := url.Parse(upstream.URL)
			if err != nil {
				logger.Error("Failed to parse upstream URL",
					slog.String("name", upstream.Name),
					slog.String("url", upstream.URL),
					slog.String("error", err.Error()),
				)
				continue
			}
			upstreams[upstream.PathPrefix] = parsedURL
			logger.Info("Registered upstream",
				slog.String("path_prefix", upstream.PathPrefix),
				slog.String("url", upstream.URL),
			)
		}
	}

	return &ReverseProxy{
		logger:    logger,
		config:    cfg,
		upstreams: upstreams,
	}
}

// ServeHTTP implements http.Handler and routes requests to appropriate upstream services
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upstreamURL, prefix := p.findUpstream(r.URL.Path)
	if upstreamURL == nil {
		p.logger.Warn("No upstream found for path",
			slog.String("path", r.URL.Path),
		)
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	// Create reverse proxy for this upstream
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

	// Set the director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Strip the path prefix before forwarding
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, prefix)
		// Update the request URI for the upstream
		req.RequestURI = req.URL.Path + req.URL.RawQuery
		if req.RequestURI == "" {
			req.RequestURI = "/"
		}
	}

	// Log the proxy request
	p.logger.Info("Proxying request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("upstream", upstreamURL.String()),
		slog.String("upstream_path", strings.TrimPrefix(r.URL.Path, prefix)),
	)

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// findUpstream finds the upstream service for a given path
func (p *ReverseProxy) findUpstream(path string) (*url.URL, string) {
	// Find the longest matching prefix
	var longestMatch string
	var matchedURL *url.URL

	for prefix, upstreamURL := range p.upstreams {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(longestMatch) {
				longestMatch = prefix
				matchedURL = upstreamURL
			}
		}
	}

	return matchedURL, longestMatch
}

// GetUpstreamHealth checks if an upstream service is healthy
func (p *ReverseProxy) GetUpstreamHealth(ctx context.Context, name string) bool {
	for _, upstream := range p.config.Upstreams {
		if upstream.Name == name {
			healthURL := upstream.URL + upstream.HealthPath
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			if err != nil {
				return false
			}

			client := &http.Client{Timeout: upstream.Timeout}
			resp, err := client.Do(req)
			if err != nil {
				p.logger.Warn("Upstream health check failed",
					slog.String("name", name),
					slog.String("url", healthURL),
					slog.String("error", err.Error()),
				)
				return false
			}
			defer resp.Body.Close()

			return resp.StatusCode >= 200 && resp.StatusCode < 300
		}
	}
	return false
}

// GetAllUpstreams returns all configured upstreams
func (p *ReverseProxy) GetAllUpstreams() []config.UpstreamConfig {
	return p.config.Upstreams
}
