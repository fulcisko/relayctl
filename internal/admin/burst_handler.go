package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

// NewBurstHandler returns an http.Handler that exposes burst balancer state
// and allows updating threshold / max-burst values at runtime.
func NewBurstHandler(bb *upstream.BurstBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetBurst(w, bb)
		case http.MethodPut:
			handleSetBurst(w, r, bb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type burstSnapshot struct {
	Threshold   int   `json:"threshold"`
	MaxBurst    int   `json:"max_burst"`
	Active      int64 `json:"active_connections"`
	InBurst     bool  `json:"in_burst"`
}

func handleGetBurst(w http.ResponseWriter, bb *upstream.BurstBalancer) {
	if bb == nil {
		http.Error(w, "burst balancer not configured", http.StatusServiceUnavailable)
		return
	}
	snap := burstSnapshot{
		Threshold: bb.Threshold(),
		MaxBurst:  bb.MaxBurst(),
		Active:    bb.Active(),
		InBurst:   bb.InBurst(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

type burstUpdateRequest struct {
	Threshold *int `json:"threshold"`
	MaxBurst  *int `json:"max_burst"`
}

func handleSetBurst(w http.ResponseWriter, r *http.Request, bb *upstream.BurstBalancer) {
	if bb == nil {
		http.Error(w, "burst balancer not configured", http.StatusServiceUnavailable)
		return
	}
	var req burstUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Threshold != nil {
		if err := bb.SetThreshold(*req.Threshold); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if req.MaxBurst != nil {
		if err := bb.SetMaxBurst(*req.MaxBurst); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	handleGetBurst(w, bb)
}
