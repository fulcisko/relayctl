package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/upstream"
)

// NewDrainHandler returns an http.Handler that exposes drain registry
// management over the admin API.
//
// GET  /admin/drain          — list all draining backends
// PUT  /admin/drain          — mark a backend as draining  {"backend":"..."}
// DELETE /admin/drain        — restore a backend            {"backend":"..."}
func NewDrainHandler(reg *upstream.DrainRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetDrain(w, reg)
		case http.MethodPut:
			handleDrainBackend(w, r, reg)
		case http.MethodDelete:
			handleRestoreBackend(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetDrain(w http.ResponseWriter, reg *upstream.DrainRegistry) {
	draining := reg.Draining()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"draining": draining})
}

func handleDrainBackend(w http.ResponseWriter, r *http.Request, reg *upstream.DrainRegistry) {
	var body struct {
		Backend string `json:"backend"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Backend == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	reg.Drain(body.Backend)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"backend": body.Backend, "status": "draining"})
}

func handleRestoreBackend(w http.ResponseWriter, r *http.Request, reg *upstream.DrainRegistry) {
	var body struct {
		Backend string `json:"backend"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Backend == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	reg.Restore(body.Backend)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"backend": body.Backend, "status": "restored"})
}
