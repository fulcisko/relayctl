package admin

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/relayctl/internal/upstream"
)

type canarySnapshot struct {
	Percent  int      `json:"percent"`
	Backends []string `json:"backends"`
}

type canaryUpdateRequest struct {
	Percent int `json:"percent"`
}

// NewCanaryHandler returns an http.Handler for inspecting and updating the
// canary traffic split at runtime.
//
// GET  /admin/canary  → returns current percent and backend list
// PUT  /admin/canary  → updates percent (body: {"percent": N})
func NewCanaryHandler(cb *upstream.CanaryBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cb == nil {
			http.Error(w, "canary balancer not configured", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetCanary(w, cb)
		case http.MethodPut:
			handleSetCanary(w, r, cb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetCanary(w http.ResponseWriter, cb *upstream.CanaryBalancer) {
	snap := canarySnapshot{
		Percent:  cb.Percent(),
		Backends: cb.Backends(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetCanary(w http.ResponseWriter, r *http.Request, cb *upstream.CanaryBalancer) {
	var req canaryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	cb.SetPercent(req.Percent)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(canarySnapshot{
		Percent:  cb.Percent(),
		Backends: cb.Backends(),
	})
}
