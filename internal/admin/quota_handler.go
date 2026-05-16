package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

type quotaRequest struct {
	Backend     string `json:"backend"`
	MaxRequests int    `json:"max_requests"`
	WindowSec   int    `json:"window_sec"`
}

// NewQuotaHandler returns an HTTP handler for managing per-backend request quotas.
func NewQuotaHandler(reg *upstream.QuotaRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetQuota(w, reg)
		case http.MethodPut:
			handleSetQuota(w, r, reg)
		case http.MethodDelete:
			handleDeleteQuota(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetQuota(w http.ResponseWriter, reg *upstream.QuotaRegistry) {
	snap := reg.Snapshot()
	type entry struct {
		MaxRequests int    `json:"max_requests"`
		WindowSec   int    `json:"window_sec"`
	}
	out := make(map[string]entry, len(snap))
	for k, v := range snap {
		out[k] = entry{
			MaxRequests: v.MaxRequests,
			WindowSec:   int(v.Window.Seconds()),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func handleSetQuota(w http.ResponseWriter, r *http.Request, reg *upstream.QuotaRegistry) {
	var req quotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	win := time.Duration(req.WindowSec) * time.Second
	if win <= 0 {
		win = time.Minute
	}
	if err := reg.Set(req.Backend, upstream.QuotaConfig{
		MaxRequests: req.MaxRequests,
		Window:      win,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteQuota(w http.ResponseWriter, r *http.Request, reg *upstream.QuotaRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
