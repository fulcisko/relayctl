package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

// PriorityConfig holds the backends→priority map used to (re)build a
// PriorityBalancer via the admin API.
type PriorityConfig struct {
	Backends map[string]int `json:"backends"`
}

// priorityHandlerState wraps the live balancer so the handler can inspect it.
type priorityHandlerState interface {
	Backends() []string
}

// NewPriorityHandler returns an http.Handler for GET /admin/priority.
// GET  – returns the ordered backend list with their positions.
// PUT  – rebuilds the balancer from a new PriorityConfig (requires hc).
func NewPriorityHandler(pb *upstream.PriorityBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetPriority(w, pb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetPriority(w http.ResponseWriter, pb *upstream.PriorityBalancer) {
	if pb == nil {
		http.Error(w, "priority balancer not configured", http.StatusServiceUnavailable)
		return
	}
	type response struct {
		Backends []string `json:"backends"`
		Count    int      `json:"count"`
	}
	backends := pb.Backends()
	resp := response{
		Backends: backends,
		Count:    len(backends),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
