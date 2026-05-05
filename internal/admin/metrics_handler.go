package admin

import (
	"encoding/json"
	"net/http"

	"github.com/user/relayctl/internal/metrics"
)

// MetricsHandler exposes collected proxy metrics over HTTP.
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a MetricsHandler backed by the given Collector.
func NewMetricsHandler(c *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{collector: c}
}

// ServeHTTP handles GET /metrics and returns a JSON snapshot.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snap := h.collector.Snapshot()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(snap); err != nil {
		http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
	}
}
