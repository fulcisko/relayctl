package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/internal/upstream"
)

// NewMirrorHandler returns an http.Handler that exposes GET / PUT endpoints
// for inspecting and toggling the MirrorBalancer's shadow mirroring state.
func NewMirrorHandler(mb *upstream.MirrorBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetMirror(w, mb)
		case http.MethodPut:
			handleSetMirror(w, r, mb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type mirrorState struct {
	Enabled  bool     `json:"enabled"`
	Primary  []string `json:"primary_backends"`
	Shadow   []string `json:"shadow_backends"`
}

func handleGetMirror(w http.ResponseWriter, mb *upstream.MirrorBalancer) {
	if mb == nil {
		http.Error(w, "mirror balancer not configured", http.StatusServiceUnavailable)
		return
	}
	state := mirrorState{
		Enabled: mb.Enabled(),
		Primary: mb.Backends(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state) //nolint:errcheck
}

type mirrorUpdateRequest struct {
	Enabled bool `json:"enabled"`
}

func handleSetMirror(w http.ResponseWriter, r *http.Request, mb *upstream.MirrorBalancer) {
	if mb == nil {
		http.Error(w, "mirror balancer not configured", http.StatusServiceUnavailable)
		return
	}
	var req mirrorUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	mb.SetEnabled(req.Enabled)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": req.Enabled}) //nolint:errcheck
}
