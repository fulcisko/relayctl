package admin

import (
	"encoding/json"
	"net/http"
)

// HealthCollector is the interface for querying backend health status.
type HealthCollector interface {
	Statuses() map[string]bool
}

// HealthCheckHandler serves health status of all registered backends.
type HealthCheckHandler struct {
	collector HealthCollector
}

// NewHealthCheckHandler creates a new HealthCheckHandler.
func NewHealthCheckHandler(c HealthCollector) *HealthCheckHandler {
	return &HealthCheckHandler{collector: c}
}

// ServeHTTP handles GET /admin/health and returns JSON backend statuses.
func (h *HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := h.collector.Statuses()

	type backendStatus struct {
		URL     string `json:"url"`
		Healthy bool   `json:"healthy"`
	}

	result := make([]backendStatus, 0, len(statuses))
	for url, healthy := range statuses {
		result = append(result, backendStatus{URL: url, Healthy: healthy})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
