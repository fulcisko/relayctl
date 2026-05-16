package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lucianoayres/relayctl/internal/upstream"
)

type headerRequest struct {
	Backend string               `json:"backend"`
	Rules   upstream.HeaderRules `json:"rules"`
}

type deleteBackendRequest struct {
	Backend string `json:"backend"`
}

// NewHeaderHandler returns an http.Handler for managing per-backend header rules.
func NewHeaderHandler(reg *upstream.HeaderRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHeaders(w, r, reg)
		case http.MethodPut:
			handleSetHeader(w, r, reg)
		case http.MethodDelete:
			handleDeleteHeader(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetHeaders(w http.ResponseWriter, _ *http.Request, reg *upstream.HeaderRegistry) {
	snapshot := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}

func handleSetHeader(w http.ResponseWriter, r *http.Request, reg *upstream.HeaderRegistry) {
	var req headerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	reg.Set(req.Backend, req.Rules)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDeleteHeader(w http.ResponseWriter, r *http.Request, reg *upstream.HeaderRegistry) {
	var req deleteBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Backend == "" {
		http.Error(w, "backend is required", http.StatusBadRequest)
		return
	}
	reg.Delete(req.Backend)
	w.WriteHeader(http.StatusNoContent)
}
