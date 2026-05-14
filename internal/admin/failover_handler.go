package admin

import (
	"encoding/json"
	"net/http"
)

// FailoverBalancer is the interface expected by NewFailoverHandler.
type FailoverBalancer interface {
	Backends() []string
	Update(backends []string) error
}

// NewFailoverHandler returns an HTTP handler for inspecting and updating
// the failover balancer's backend priority list.
func NewFailoverHandler(fb FailoverBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetFailover(w, fb)
		case http.MethodPut:
			handleSetFailover(w, r, fb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type failoverResponse struct {
	Backends []string `json:"backends"`
}

func handleGetFailover(w http.ResponseWriter, fb FailoverBalancer) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(failoverResponse{Backends: fb.Backends()})
}

func handleSetFailover(w http.ResponseWriter, r *http.Request, fb FailoverBalancer) {
	var req failoverResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := fb.Update(req.Backends); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(failoverResponse{Backends: fb.Backends()})
}
