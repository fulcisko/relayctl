package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/upstream"
)

// NewRateLimitRegistryHandler returns an http.Handler for managing per-backend
// rate limit configurations via GET / PUT / DELETE.
func NewRateLimitRegistryHandler(reg *upstream.RateLimitRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetRateLimitRegistry(w, reg)
		case http.MethodPut:
			handleSetRateLimitRegistry(w, r, reg)
		case http.MethodDelete:
			handleDeleteRateLimitRegistry(w, r, reg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetRateLimitRegistry(w http.ResponseWriter, reg *upstream.RateLimitRegistry) {
	snap := reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func handleSetRateLimitRegistry(w http.ResponseWriter, r *http.Request, reg *upstream.RateLimitRegistry) {
	var body struct {
		Backend           string  `json:"backend"`
		RequestsPerSecond float64 `json:"requests_per_second"`
		Burst             int     `json:"burst"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	cfg := upstream.PerBackendRateLimit{
		RequestsPerSecond: body.RequestsPerSecond,
		Burst:             body.Burst,
	}
	if err := reg.Set(body.Backend, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteRateLimitRegistry(w http.ResponseWriter, r *http.Request, reg *upstream.RateLimitRegistry) {
	backend := r.URL.Query().Get("backend")
	if backend == "" {
		http.Error(w, "missing backend query param", http.StatusBadRequest)
		return
	}
	reg.Delete(backend)
	w.WriteHeader(http.StatusNoContent)
}
