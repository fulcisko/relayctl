package admin

import (
	"encoding/json"
	"net/http"

	"github.com/user/relayctl/internal/retry"
)

// RetryPolicyStore allows reading and updating the active retry policy.
type RetryPolicyStore interface {
	GetPolicy() retry.Policy
	SetPolicy(retry.Policy)
}

// NewRetryHandler returns an http.Handler that exposes GET/PUT endpoints
// for inspecting and updating the retry policy at runtime.
func NewRetryHandler(store RetryPolicyStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetPolicy(w, store)
		case http.MethodPut:
			handleSetPolicy(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGetPolicy(w http.ResponseWriter, store RetryPolicyStore) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.GetPolicy()) //nolint:errcheck
}

func handleSetPolicy(w http.ResponseWriter, r *http.Request, store RetryPolicyStore) {
	var p retry.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if p.MaxAttempts < 1 {
		http.Error(w, "maxAttempts must be >= 1", http.StatusUnprocessableEntity)
		return
	}
	store.SetPolicy(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.GetPolicy()) //nolint:errcheck
}
