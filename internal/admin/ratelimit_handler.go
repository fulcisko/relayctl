package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/relayctl/internal/ratelimit"
)

// RateLimitConfig holds configuration for the rate limiter exposed via admin API.
type RateLimitConfig struct {
	Max    int           `json:"max"`
	Window time.Duration `json:"window"`
}

// RateLimitHandler manages the admin endpoint for rate limit status.
type RateLimitHandler struct {
	limiter *ratelimit.Limiter
	cfg     RateLimitConfig
}

// NewRateLimitHandler creates a new RateLimitHandler.
func NewRateLimitHandler(l *ratelimit.Limiter, cfg RateLimitConfig) *RateLimitHandler {
	return &RateLimitHandler{limiter: l, cfg: cfg}
}

// ServeHTTP handles GET /admin/ratelimit — returns current rate limit config.
func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := map[string]interface{}{
		"max":           h.cfg.Max,
		"window_seconds": h.cfg.Window.Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
