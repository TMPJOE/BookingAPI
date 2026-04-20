package health

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"hotel.com/bookingapi/internal/config"
	"hotel.com/bookingapi/internal/logging"
)

// UpstreamHealth represents the health status of an upstream service
type UpstreamHealth struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Healthy   bool      `json:"healthy"`
	CheckedAt time.Time `json:"checked_at"`
	Error     string    `json:"error,omitempty"`
}

// Checker monitors the health of upstream services
type Checker struct {
	logger     *logging.Logger
	config     *config.Config
	status     map[string]*UpstreamHealth
	statusMu   sync.RWMutex
	stopCh     chan struct{}
	interval   time.Duration
	httpClient *http.Client
}

// NewChecker creates a new health checker
func NewChecker(logger *logging.Logger, cfg *config.Config, interval time.Duration) *Checker {
	return &Checker{
		logger:   logger,
		config:   cfg,
		status:   make(map[string]*UpstreamHealth),
		stopCh:   make(chan struct{}),
		interval: interval,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start begins periodic health checks for all upstreams
func (c *Checker) Start() {
	c.logger.Info("Starting upstream health checker",
		slog.String("interval", c.interval.String()),
	)

	// Run initial check
	c.checkAll()

	// Start periodic checks
	ticker := time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.checkAll()
			case <-c.stopCh:
				ticker.Stop()
				c.logger.Info("Upstream health checker stopped")
				return
			}
		}
	}()
}

// Stop stops the health checker
func (c *Checker) Stop() {
	close(c.stopCh)
}

// checkAll checks health of all upstream services
func (c *Checker) checkAll() {
	var wg sync.WaitGroup
	for _, upstream := range c.config.Upstreams {
		wg.Add(1)
		go func(u config.UpstreamConfig) {
			defer wg.Done()
			c.checkUpstream(u)
		}(upstream)
	}
	wg.Wait()
}

// checkUpstream checks the health of a single upstream service
func (c *Checker) checkUpstream(upstream config.UpstreamConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), upstream.Timeout)
	defer cancel()

	healthURL := upstream.URL + upstream.HealthPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		c.updateStatus(upstream.Name, upstream.URL, false, err.Error())
		return
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.updateStatus(upstream.Name, upstream.URL, false, err.Error())
		c.logger.Warn("Upstream health check failed",
			slog.String("name", upstream.Name),
			slog.String("url", healthURL),
			slog.String("error", err.Error()),
		)
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	if healthy {
		c.updateStatus(upstream.Name, upstream.URL, true, "")
		c.logger.Debug("Upstream health check passed",
			slog.String("name", upstream.Name),
			slog.Int("status_code", resp.StatusCode),
		)
	} else {
		c.updateStatus(upstream.Name, upstream.URL, false, "unhealthy status code")
		c.logger.Warn("Upstream health check returned unhealthy status",
			slog.String("name", upstream.Name),
			slog.Int("status_code", resp.StatusCode),
		)
	}
}

// updateStatus updates the health status of an upstream
func (c *Checker) updateStatus(name, url string, healthy bool, errorMsg string) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()

	c.status[name] = &UpstreamHealth{
		Name:      name,
		URL:       url,
		Healthy:   healthy,
		CheckedAt: time.Now(),
		Error:     errorMsg,
	}
}

// GetStatus returns the health status of all upstreams
func (c *Checker) GetStatus() []UpstreamHealth {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()

	result := make([]UpstreamHealth, 0, len(c.status))
	for _, h := range c.status {
		result = append(result, *h)
	}
	return result
}

// IsHealthy returns true if the specified upstream is healthy
func (c *Checker) IsHealthy(name string) bool {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()

	if h, ok := c.status[name]; ok {
		return h.Healthy
	}
	return false
}

// GetHealthyCount returns the count of healthy upstreams
func (c *Checker) GetHealthyCount() int {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()

	count := 0
	for _, h := range c.status {
		if h.Healthy {
			count++
		}
	}
	return count
}

// GetTotalCount returns the total count of configured upstreams
func (c *Checker) GetTotalCount() int {
	return len(c.config.Upstreams)
}
