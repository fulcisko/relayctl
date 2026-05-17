package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lukedever/relayctl/internal/upstream"
)

type hedgeRequest struct {
	Backend   string `json:"backend"`
	DelayMS   int64  `json:"delay_ms"`
	MaxHedges int    `json:"max_hedges"`
}

// NewHedgeHandler returns an http.Handler for managing the HedgeRegistry.
//
//	GET  /admin/hedge          — snapshot of all entries
//	PUT  /admin/hedge          — set a hedge config for a backend
//	DELETE /admin/hedge?backend=<url> — remove a backend's config
func NewHedgeHandler(reg *upstream.HedgeRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHedge(w, reg)
		case http.MethodPut:
			handleSetHedge(w, r, reg)
		case http.MethodDelete:
			handleDeleteHedge(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetHedge(w http.ResponseWriter, reg *upstream.HedgeRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetHedge(w http.ResponseWriter, r *http.Request, reg *upstream.HedgeRegistry) {
	var req hedgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Backend == "" {
		http.Error(w, "backend required", http.StatusBadRequest)
		return
	}
	cfg := upstream.HedgeConfig{
		Delay:     time.Duration(req.DelayMS) * time.Millisecond,
		MaxHedges: req.MaxHedges,
	}
	if err := reg.Set(req.Backend, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteHedge(w http.ResponseWriter, r *http.Request, reg *upstream.HedgeRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
