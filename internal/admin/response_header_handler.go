package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

// NewResponseHeaderHandler returns an http.Handler for managing per-backend
// response header rules via the admin API.
//
// GET  /admin/response-headers          — list all rules
// PUT  /admin/response-headers          — set rules for a backend (JSON body)
// DELETE /admin/response-headers?backend=<url> — remove rules for a backend
func NewResponseHeaderHandler(reg *upstream.ResponseHeaderRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetResponseHeaders(w, reg)
		case http.MethodPut:
			handleSetResponseHeader(w, r, reg)
		case http.MethodDelete:
			handleDeleteResponseHeader(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetResponseHeaders(w http.ResponseWriter, reg *upstream.ResponseHeaderRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetResponseHeader(w http.ResponseWriter, r *http.Request, reg *upstream.ResponseHeaderRegistry) {
	var payload struct {
		Backend string                          `json:"backend"`
		Rules   upstream.ResponseHeaderRules    `json:"rules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	if err := reg.Set(payload.Backend, payload.Rules); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteResponseHeader(w http.ResponseWriter, r *http.Request, reg *upstream.ResponseHeaderRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
