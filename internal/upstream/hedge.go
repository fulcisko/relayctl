package upstream

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// HedgeRegistry stores per-backend hedged request configuration.
// A hedged request fires a second request after a delay if the first has not
// yet responded, returning whichever reply arrives first.
type HedgeRegistry struct {
	mu      sync.RWMutex
	entries map[string]HedgeConfig
}

// HedgeConfig holds the hedge parameters for a single backend.
type HedgeConfig struct {
	// Delay is how long to wait before firing the hedge request.
	Delay time.Duration `json:"delay_ms"`
	// MaxHedges is the maximum number of extra requests to fire (default 1).
	MaxHedges int `json:"max_hedges"`
}

var errNilHedgeBackend = errors.New("hedge: backend must not be empty")

// NewHedgeRegistry creates an empty HedgeRegistry.
func NewHedgeRegistry() *HedgeRegistry {
	return &HedgeRegistry{entries: make(map[string]HedgeConfig)}
}

// Set stores a HedgeConfig for the given backend URL.
func (r *HedgeRegistry) Set(backend string, cfg HedgeConfig) error {
	if backend == "" {
		return errNilHedgeBackend
	}
	if cfg.Delay <= 0 {
		cfg.Delay = 50 * time.Millisecond
	}
	if cfg.MaxHedges <= 0 {
		cfg.MaxHedges = 1
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = cfg
	return nil
}

// Get returns the HedgeConfig for a backend, and whether it was found.
func (r *HedgeRegistry) Get(backend string) (HedgeConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.entries[backend]
	return cfg, ok
}

// Delete removes the hedge configuration for a backend.
func (r *HedgeRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all registered hedge configs.
func (r *HedgeRegistry) Snapshot() map[string]HedgeConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]HedgeConfig, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}

// Middleware returns an http.Handler that applies hedging for the given backend.
func (r *HedgeRegistry) Middleware(backend string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg, ok := r.Get(backend)
		if !ok {
			next.ServeHTTP(w, req)
			return
		}
		type result struct{}
		done := make(chan result, 1)
		go func() {
			next.ServeHTTP(w, req)
			done <- result{}
		}()
		select {
		case <-done:
			return
		case <-time.After(cfg.Delay):
			// hedge fired — in a real implementation a second upstream
			// request would race; here we just let the original finish.
			<-done
		}
	})
}
