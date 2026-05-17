package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/luketherrien/relayctl/internal/upstream"
)

// NewDeadlineHandler returns an HTTP handler for managing per-backend deadlines.
//
// GET  /admin/deadline?backend=<url>  — retrieve deadline for a backend
// PUT  /admin/deadline                — set deadline (JSON body: {"backend": "...", "deadline_ms": 5000})
// DELETE /admin/deadline?backend=<url> — remove deadline for a backend
func NewDeadlineHandler(reg *upstream.DeadlineRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetDeadline(w, r, reg)
		case http.MethodPut:
			handleSetDeadline(w, r, reg)
		case http.MethodDelete:
			handleDeleteDeadline(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetDeadline(w http.ResponseWriter, r *http.Request, reg *upstream.DeadlineRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		snap := reg.Snapshot()
		out := make(map[string]int64, len(snap))
		for k, v := range snap {
			out[k] = v.Milliseconds()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	}
	d, ok := reg.Get(backend)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"deadline_ms": d.Milliseconds()})
}

func handleSetDeadline(w http.ResponseWriter, r *http.Request, reg *upstream.DeadlineRegistry) {
	var body struct {
		Backend    string `json:"backend"`
		DeadlineMS int64  `json:"deadline_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := reg.Set(body.Backend, time.Duration(body.DeadlineMS)*time.Millisecond); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteDeadline(w http.ResponseWriter, r *http.Request, reg *upstream.DeadlineRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
