package admin

import (
	"encoding/json"
	"net/http"

	"github.com/user/relayctl/internal/upstream"
)

// BalancerHandler exposes the upstream balancer state via the admin API.
type BalancerHandler struct {
	balancer *upstream.Balancer
}

// NewBalancerHandler creates a new BalancerHandler.
func NewBalancerHandler(b *upstream.Balancer) *BalancerHandler {
	return &BalancerHandler{balancer: b}
}

func (h *BalancerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPut:
		h.handleUpdate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *BalancerHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	type response struct {
		Backends []string `json:"backends"`
		Count    int      `json:"count"`
	}
	backends := h.balancer.Backends()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Backends: backends, Count: len(backends)})
}

func (h *BalancerHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Backends []string `json:"backends"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.balancer.Update(req.Backends); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
