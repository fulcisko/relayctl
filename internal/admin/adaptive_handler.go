package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/relayctl/internal/upstream"
)

// adaptiveSnapshot holds the latency snapshot for a single backend.
type adaptiveSnapshot struct {
	Backend    string        `json:"backend"`
	AvgLatency time.Duration `json:"avg_latency_ns"`
	Samples    int           `json:"samples"`
}

// NewAdaptiveHandler returns an HTTP handler that exposes the adaptive
// balancer's per-backend latency statistics as JSON.
func NewAdaptiveHandler(b *upstream.AdaptiveBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if b == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
			return
		}

		backends := b.Backends()
		snaps := make([]adaptiveSnapshot, 0, len(backends))
		for _, addr := range backends {
			snaps = append(snaps, adaptiveSnapshot{
				Backend:    addr,
				AvgLatency: b.AvgLatency(addr),
				Samples:    0, // samples not exported; latency sufficient for display
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snaps)
	})
}
