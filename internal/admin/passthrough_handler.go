package admin

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

// NewPassthroughHandler returns an HTTP handler that exposes the
// PassthroughRegistry over a simple REST API.
//
//	GET  /admin/passthrough          — list all passthrough backends
//	PUT  /admin/passthrough          — register a backend as passthrough
//	DELETE /admin/passthrough        — remove a backend from passthrough
func NewPassthroughHandler(reg *upstream.PassthroughRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetPassthrough(w, reg)
		case http.MethodPut:
			handleSetPassthrough(w, r, reg)
		case http.MethodDelete:
			handleDeletePassthrough(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetPassthrough(w http.ResponseWriter, reg *upstream.PassthroughRegistry) {
	snap := reg.Snapshot()
	sort.Strings(snap)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"backends": snap})
}

func handleSetPassthrough(w http.ResponseWriter, r *http.Request, reg *upstream.PassthroughRegistry) {
	var body struct {
		Backend string `json:"backend"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeletePassthrough(w http.ResponseWriter, r *http.Request, reg *upstream.PassthroughRegistry) {
	var body struct {
		Backend string `json:"backend"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	reg.Delete(body.Backend)
	w.WriteHeader(http.StatusNoContent)
}
