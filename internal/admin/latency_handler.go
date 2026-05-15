package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

// latencySnapshot is the JSON shape returned by the handler.
type latencySnapshot struct {
	Backends map[string]string `json:"backends"`
}

// NewLatencyHandler returns an HTTP handler that exposes per-backend
// average latency from the provided LatencyTracker.
func NewLatencyHandler(tracker *upstream.LatencyTracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if tracker == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(latencySnapshot{Backends: map[string]string{}})
			return
		}
		raw := tracker.Snapshot()
		formatted := make(map[string]string, len(raw))
		for backend, avg := range raw {
			formatted[backend] = formatDuration(avg)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(latencySnapshot{Backends: formatted})
	})
}

// formatDuration renders a duration as a human-readable string with ms precision.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0ms"
	}
	return d.Round(time.Microsecond).String()
}
