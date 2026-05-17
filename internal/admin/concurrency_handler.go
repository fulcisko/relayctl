package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukasmalkmus/relayctl/internal/upstream"
)

// NewConcurrencyHandler returns an HTTP handler for managing per-backend
// concurrency limits.
//
// GET  /admin/concurrency?backend=<url>  — retrieve limit and active count
// PUT  /admin/concurrency               — set limit {"backend":"...","max":N}
// DELETE /admin/concurrency?backend=<url> — remove limit
func NewConcurrencyHandler(reg *upstream.ConcurrencyRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetConcurrency(w, r, reg)
		case http.MethodPut:
			handleSetConcurrency(w, r, reg)
		case http.MethodDelete:
			handleDeleteConcurrency(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetConcurrency(w http.ResponseWriter, r *http.Request, reg *upstream.ConcurrencyRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	max, active, ok := reg.Get(backend)
	if !ok {
		http.Error(w, "backend not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"max": max, "active": active})
}

func handleSetConcurrency(w http.ResponseWriter, r *http.Request, reg *upstream.ConcurrencyRegistry) {
	var body struct {
		Backend string `json:"backend"`
		Max     int64  `json:"max"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, body.Max); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteConcurrency(w http.ResponseWriter, r *http.Request, reg *upstream.ConcurrencyRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
