package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lucianoayres/relayctl/internal/circuitbreaker"
)

// CircuitBreakerRegistryHandler exposes the circuit breaker registry via HTTP.
type CircuitBreakerRegistryHandler struct {
	registry *circuitbreaker.Registry
}

// NewCircuitBreakerRegistryHandler creates a new handler backed by the given registry.
func NewCircuitBreakerRegistryHandler(r *circuitbreaker.Registry) *CircuitBreakerRegistryHandler {
	return &CircuitBreakerRegistryHandler{registry: r}
}

func (h *CircuitBreakerRegistryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleSnapshot(w, r)
	case http.MethodDelete:
		h.handleReset(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CircuitBreakerRegistryHandler) handleSnapshot(w http.ResponseWriter, _ *http.Request) {
	snap := h.registry.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap) //nolint:errcheck
}

func (h *CircuitBreakerRegistryHandler) handleReset(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("backend")
	if key == "" {
		http.Error(w, "missing 'backend' query parameter", http.StatusBadRequest)
		return
	}
	h.registry.Reset(key)
	w.WriteHeader(http.StatusNoContent)
}
