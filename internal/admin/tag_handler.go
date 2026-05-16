package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

// NewTagHandler returns an HTTP handler for managing backend tag registries.
// GET  /admin/tags          — returns snapshot of all tags
// PUT  /admin/tags          — sets tags for a backend {"backend":"...","tags":[...]}
// DELETE /admin/tags?backend=... — removes tags for a backend
func NewTagHandler(reg *upstream.TagRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTags(w, reg)
		case http.MethodPut:
			handleSetTag(w, r, reg)
		case http.MethodDelete:
			handleDeleteTag(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetTags(w http.ResponseWriter, reg *upstream.TagRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetTag(w http.ResponseWriter, r *http.Request, reg *upstream.TagRegistry) {
	var body struct {
		Backend string   `json:"backend"`
		Tags    []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, body.Tags); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteTag(w http.ResponseWriter, r *http.Request, reg *upstream.TagRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query parameter", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
