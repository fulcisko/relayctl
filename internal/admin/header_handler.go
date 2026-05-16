package admin

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/relayctl/internal/upstream"
)

// NewHeaderHandler returns an http.Handler for managing per-backend header rules.
//
// GET  /admin/headers          → snapshot of all rules
// PUT  /admin/headers          → set rule for a backend  {"backend":"...", "rule":{...}}
// DELETE /admin/headers?backend=<url> → remove rule
func NewHeaderHandler(reg *upstream.HeaderRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHeaders(w, reg)
		case http.MethodPut:
			handleSetHeader(w, r, reg)
		case http.MethodDelete:
			handleDeleteHeader(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetHeaders(w http.ResponseWriter, reg *upstream.HeaderRegistry) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reg.Snapshot())
}

func handleSetHeader(w http.ResponseWriter, r *http.Request, reg *upstream.HeaderRegistry) {
	var body struct {
		Backend string             `json:"backend"`
		Rule    upstream.HeaderRule `json:"rule"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, body.Rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteHeader(w http.ResponseWriter, r *http.Request, reg *upstream.HeaderRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
