package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/radovskyb/relayctl/internal/upstream"
)

// NewBackoffHandler returns an HTTP handler for managing per-backend backoff policies.
// GET  /admin/backoff          — list all policies
// PUT  /admin/backoff          — set a policy (JSON body)
// DELETE /admin/backoff?backend=<url> — remove a policy
func NewBackoffHandler(reg *upstream.BackoffRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetBackoff(w, reg)
		case http.MethodPut:
			handleSetBackoff(w, r, reg)
		case http.MethodDelete:
			handleDeleteBackoff(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetBackoff(w http.ResponseWriter, reg *upstream.BackoffRegistry) {
	snap := reg.Snapshot()
	type policyView struct {
		BaseDelayMS  int64   `json:"base_delay_ms"`
		MaxDelayMS   int64   `json:"max_delay_ms"`
		Multiplier   float64 `json:"multiplier"`
		Jitter       float64 `json:"jitter"`
	}
	out := make(map[string]policyView, len(snap))
	for k, p := range snap {
		out[k] = policyView{
			BaseDelayMS: p.BaseDelay.Milliseconds(),
			MaxDelayMS:  p.MaxDelay.Milliseconds(),
			Multiplier:  p.Multiplier,
			Jitter:      p.Jitter,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func handleSetBackoff(w http.ResponseWriter, r *http.Request, reg *upstream.BackoffRegistry) {
	var body struct {
		Backend     string  `json:"backend"`
		BaseDelayMS int64   `json:"base_delay_ms"`
		MaxDelayMS  int64   `json:"max_delay_ms"`
		Multiplier  float64 `json:"multiplier"`
		Jitter      float64 `json:"jitter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	p := upstream.BackoffPolicy{
		BaseDelay:  time.Duration(body.BaseDelayMS) * time.Millisecond,
		MaxDelay:   time.Duration(body.MaxDelayMS) * time.Millisecond,
		Multiplier: body.Multiplier,
		Jitter:     body.Jitter,
	}
	if err := reg.Set(body.Backend, p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteBackoff(w http.ResponseWriter, r *http.Request, reg *upstream.BackoffRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "backend query param required", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
