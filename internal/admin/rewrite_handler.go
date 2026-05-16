package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/internal/upstream"
)

// NewRewriteHandler returns an http.Handler for managing path rewrite rules.
//
// GET  /admin/rewrite          — list all rules
// PUT  /admin/rewrite          — set a rule  {"backend":"...","prefix":"...","replacement":"..."}
// DELETE /admin/rewrite?backend=... — remove a rule
func NewRewriteHandler(reg *upstream.RewriteRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reg == nil {
			http.Error(w, "rewrite registry unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetRewrite(w, reg)
		case http.MethodPut:
			handleSetRewrite(w, r, reg)
		case http.MethodDelete:
			handleDeleteRewrite(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetRewrite(w http.ResponseWriter, reg *upstream.RewriteRegistry) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reg.Snapshot())
}

func handleSetRewrite(w http.ResponseWriter, r *http.Request, reg *upstream.RewriteRegistry) {
	var body struct {
		Backend     string `json:"backend"`
		Prefix      string `json:"prefix"`
		Replacement string `json:"replacement"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, body.Prefix, body.Replacement); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteRewrite(w http.ResponseWriter, r *http.Request, reg *upstream.RewriteRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
