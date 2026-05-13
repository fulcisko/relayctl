package admin

import (
	"encoding/json"
	"net/http"
	"sync"
)

// AccessLogCollector stores recent access log entries in a ring buffer.
type AccessLogCollector struct {
	mu      sync.RWMutex
	entries []map[string]any
	max     int
}

// NewAccessLogCollector creates a collector that retains up to max entries.
func NewAccessLogCollector(max int) *AccessLogCollector {
	if max <= 0 {
		max = 100
	}
	return &AccessLogCollector{max: max}
}

// Write implements io.Writer so it can be passed to accesslog.New.
func (c *AccessLogCollector) Write(p []byte) (int, error) {
	var entry map[string]any
	if err := json.Unmarshal(p, &entry); err != nil {
		return len(p), nil // silently skip malformed lines
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		c.entries = c.entries[1:]
	}
	c.entries = append(c.entries, entry)
	return len(p), nil
}

// Snapshot returns a copy of the current entries.
func (c *AccessLogCollector) Snapshot() []map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]map[string]any, len(c.entries))
	copy(out, c.entries)
	return out
}

// NewAccessLogHandler returns an HTTP handler exposing recent access log entries.
func NewAccessLogHandler(c *AccessLogCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		entries := c.Snapshot()
		if entries == nil {
			entries = []map[string]any{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"count":   len(entries),
			"entries": entries,
		})
	})
}
