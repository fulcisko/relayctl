package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lukethinker/relayctl/internal/upstream"
)

// NewCircuitPolicyHandler returns an HTTP handler for managing per-backend
// circuit breaker policies via the admin API.
func NewCircuitPolicyHandler(reg *upstream.CircuitPolicyRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetCircuitPolicy(w, r, reg)
		case http.MethodPut:
			handleSetCircuitPolicy(w, r, reg)
		case http.MethodDelete:
			handleDeleteCircuitPolicy(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetCircuitPolicy(w http.ResponseWriter, _ *http.Request, reg *upstream.CircuitPolicyRegistry) {
	snapshot := reg.Snapshot()
	type entry struct {
		Backend   string `json:"backend"`
		Threshold int    `json:"threshold"`
		Timeout   string `json:"timeout"`
	}
	result := make([]entry, 0, len(snapshot))
	for backend, p := range snapshot {
		result = append(result, entry{
			Backend:   backend,
			Threshold: p.Threshold,
			Timeout:   p.Timeout.String(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"policies": result})
}

func handleSetCircuitPolicy(w http.ResponseWriter, r *http.Request, reg *upstream.CircuitPolicyRegistry) {
	var req struct {
		Backend   string `json:"backend"`
		Threshold int    `json:"threshold"`
		Timeout   string `json:"timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		http.Error(w, "invalid timeout duration", http.StatusBadRequest)
		return
	}
	if err := reg.Set(req.Backend, upstream.CircuitPolicy{
		Threshold: req.Threshold,
		Timeout:   timeout,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDeleteCircuitPolicy(w http.ResponseWriter, r *http.Request, reg *upstream.CircuitPolicyRegistry) {
	var req struct {
		Backend string `json:"backend"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	reg.Delete(req.Backend)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
