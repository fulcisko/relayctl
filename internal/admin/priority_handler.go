package admin

import (
	"encoding/json"
	"net/http"

	"github.com/lukethinnes/relayctl/internal/upstream"
)

// NewPriorityHandler returns an HTTP handler that exposes the current
// priority balancer configuration via GET /admin/priority.
func NewPriorityHandler(b *upstream.PriorityBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if b == nil {
			http.Error(w, "priority balancer not configured", http.StatusServiceUnavailable)
			return
		}

		groups := b.Groups()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Groups []upstream.PriorityGroup `json:"groups"`
		}{
			Groups: groups,
		})
	})
}
