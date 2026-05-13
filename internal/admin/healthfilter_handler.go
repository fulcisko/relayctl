package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

// HealthFilterHandler serves healthy backend information from a FilteredBalancer.
type HealthFilterHandler struct {
	fb *upstream.FilteredBalancer
}

// NewHealthFilterHandler creates a new HealthFilterHandler.
func NewHealthFilterHandler(fb *upstream.FilteredBalancer) *HealthFilterHandler {
	return &HealthFilterHandler{fb: fb}
}

type healthFilterResponse struct {
	Healthy []string `json:"healthy"`
	Count   int      `json:"count"`
}

func (h *HealthFilterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.fb == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthFilterResponse{Healthy: []string{}, Count: 0})
		return
	}

	backends := h.fb.HealthyBackends()
	if backends == nil {
		backends = []string{}
	}

	resp := healthFilterResponse{
		Healthy: backends,
		Count:   len(backends),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
