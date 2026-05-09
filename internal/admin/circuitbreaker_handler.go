package admin

import (
	"encoding/json"
	"net/http"

	"github.com/user/relayctl/internal/circuitbreaker"
)

// NewCircuitBreakerHandler returns an HTTP handler that exposes circuit breaker
// state and allows resetting it via the admin API.
func NewCircuitBreakerHandler(cb *circuitbreaker.CircuitBreaker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetCBState(w, cb)
		case http.MethodPost:
			handleResetCB(w, r, cb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetCBState(w http.ResponseWriter, cb *circuitbreaker.CircuitBreaker) {
	snap := cb.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(snap)
}

func handleResetCB(w http.ResponseWriter, r *http.Request, cb *circuitbreaker.CircuitBreaker) {
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Action != "reset" {
		http.Error(w, "unsupported action, use \"reset\"", http.StatusBadRequest)
		return
	}
	cb.Reset()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}
