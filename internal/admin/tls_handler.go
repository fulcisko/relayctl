package admin

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/relayctl/internal/upstream"
)

// NewTLSHandler returns an http.Handler for managing per-backend TLS overrides.
//
// GET  /admin/tls          → snapshot of all registered TLS configs
// PUT  /admin/tls          → set TLS config for a backend
// DELETE /admin/tls?backend=<url> → remove TLS config
func NewTLSHandler(reg *upstream.TLSRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTLS(w, reg)
		case http.MethodPut:
			handleSetTLS(w, r, reg)
		case http.MethodDelete:
			handleDeleteTLS(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetTLS(w http.ResponseWriter, reg *upstream.TLSRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

type tlsSetRequest struct {
	Backend            string `json:"backend"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	ServerName         string `json:"server_name"`
}

func handleSetTLS(w http.ResponseWriter, r *http.Request, reg *upstream.TLSRegistry) {
	var req tlsSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(req.Backend, upstream.TLSConfig{
		InsecureSkipVerify: req.InsecureSkipVerify,
		ServerName:         req.ServerName,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteTLS(w http.ResponseWriter, r *http.Request, reg *upstream.TLSRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
