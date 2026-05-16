package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/internal/upstream"
)

// NewShadowingHandler returns an http.Handler for managing the ShadowRegistry.
// GET  /admin/shadowing          – list all entries
// PUT  /admin/shadowing          – set an entry   {"backend":"...", "shadow":{...}}
// DELETE /admin/shadowing?backend=... – remove an entry
func NewShadowingHandler(reg *upstream.ShadowRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetShadowing(w, reg)
		case http.MethodPut:
			handleSetShadowing(w, r, reg)
		case http.MethodDelete:
			handleDeleteShadowing(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetShadowing(w http.ResponseWriter, reg *upstream.ShadowRegistry) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reg.Snapshot())
}

func handleSetShadowing(w http.ResponseWriter, r *http.Request, reg *upstream.ShadowRegistry) {
	var body struct {
		Backend string               `json:"backend"`
		Shadow  upstream.ShadowEntry `json:"shadow"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Backend == "" {
		http.Error(w, "missing backend", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, body.Shadow); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteShadowing(w http.ResponseWriter, r *http.Request, reg *upstream.ShadowRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
