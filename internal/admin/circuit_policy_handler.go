package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/upstream"
)

// NewCircuitPolicyHandler returns an http.Handler for managing per-backend
// circuit breaker policies via GET, PUT, and DELETE.
//
//	GET    /admin/circuit-policy          → snapshot of all policies
//	PUT    /admin/circuit-policy?backend= → set policy for backend
//	DELETE /admin/circuit-policy?backend= → remove policy for backend
func NewCircuitPolicyHandler(reg *upstream.CircuitPolicyRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetCircuitPolicy(w, reg)
		case http.MethodPut:
			handleSetCircuitPolicy(w, r, reg)
		case http.MethodDelete:
			handleDeleteCircuitPolicy(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetCircuitPolicy(w http.ResponseWriter, reg *upstream.CircuitPolicyRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetCircuitPolicy(w http.ResponseWriter, r *http.Request, reg *upstream.CircuitPolicyRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	var p upstream.CircuitPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := reg.Set(backend, p); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteCircuitPolicy(w http.ResponseWriter, r *http.Request, reg *upstream.CircuitPolicyRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
