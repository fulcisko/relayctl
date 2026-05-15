package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/relayctl/internal/upstream"
)

// timeoutPayload is the JSON shape for timeout API requests and responses.
type timeoutPayload struct {
	Backend        string `json:"backend"`
	DialMS         int64  `json:"dial_ms"`
	ResponseHeaderMS int64 `json:"response_header_ms"`
	IdleMS         int64  `json:"idle_ms"`
}

// NewTimeoutHandler returns an http.Handler for the /admin/timeouts endpoint.
// GET  /admin/timeouts          — list all backend timeout configs
// PUT  /admin/timeouts          — set timeout config for a backend
// DELETE /admin/timeouts?backend=<url> — remove config for a backend
func NewTimeoutHandler(reg *upstream.TimeoutRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTimeouts(w, reg)
		case http.MethodPut:
			handleSetTimeout(w, r, reg)
		case http.MethodDelete:
			handleDeleteTimeout(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetTimeouts(w http.ResponseWriter, reg *upstream.TimeoutRegistry) {
	snap := reg.Snapshot()
	out := make([]timeoutPayload, 0, len(snap))
	for backend, cfg := range snap {
		out = append(out, timeoutPayload{
			Backend:          backend,
			DialMS:           cfg.Dial.Milliseconds(),
			ResponseHeaderMS: cfg.ResponseHeader.Milliseconds(),
			IdleMS:           cfg.Idle.Milliseconds(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func handleSetTimeout(w http.ResponseWriter, r *http.Request, reg *upstream.TimeoutRegistry) {
	var p timeoutPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	cfg := upstream.TimeoutConfig{
		Dial:           time.Duration(p.DialMS) * time.Millisecond,
		ResponseHeader: time.Duration(p.ResponseHeaderMS) * time.Millisecond,
		Idle:           time.Duration(p.IdleMS) * time.Millisecond,
	}
	if err := reg.Set(p.Backend, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteTimeout(w http.ResponseWriter, r *http.Request, reg *upstream.TimeoutRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
