package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/relayctl/relayctl/internal/upstream"
)

// NewWarmupHandler returns an http.Handler for managing backend warmup.
//
// GET  /admin/warmup          → snapshot of all backend weights
// PUT  /admin/warmup          → register a backend for warmup
// DELETE /admin/warmup?backend=<url> → remove backend from warmup
func NewWarmupHandler(reg *upstream.WarmupRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetWarmup(w, reg)
		case http.MethodPut:
			handleSetWarmup(w, r, reg)
		case http.MethodDelete:
			handleDeleteWarmup(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetWarmup(w http.ResponseWriter, reg *upstream.WarmupRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

type warmupRequest struct {
	Backend      string `json:"backend"`
	MaxWeight    int    `json:"max_weight"`
	RampSeconds  int    `json:"ramp_seconds"`
}

func handleSetWarmup(w http.ResponseWriter, r *http.Request, reg *upstream.WarmupRegistry) {
	var req warmupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	cfg := upstream.WarmupConfig{
		MaxWeight:    req.MaxWeight,
		RampDuration: time.Duration(req.RampSeconds) * time.Second,
	}
	if err := reg.Register(req.Backend, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteWarmup(w http.ResponseWriter, r *http.Request, reg *upstream.WarmupRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
