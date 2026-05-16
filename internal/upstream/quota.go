package upstream

import (
	"errors"
	"sync"
	"time"
)

// QuotaConfig defines a rolling-window request quota for a backend.
type QuotaConfig struct {
	MaxRequests int           // maximum requests allowed in the window
	Window      time.Duration // rolling window duration
}

type quotaState struct {
	cfg      QuotaConfig
	count    int
	windowAt time.Time
}

// QuotaRegistry tracks per-backend request quotas using a rolling window.
type QuotaRegistry struct {
	mu      sync.Mutex
	entries map[string]*quotaState
}

// NewQuotaRegistry returns an empty QuotaRegistry.
func NewQuotaRegistry() *QuotaRegistry {
	return &QuotaRegistry{entries: make(map[string]*quotaState)}
}

// Set registers or updates the quota configuration for a backend.
func (r *QuotaRegistry) Set(backend string, cfg QuotaConfig) error {
	if backend == "" {
		return errors.New("quota: backend must not be empty")
	}
	if cfg.MaxRequests <= 0 {
		return errors.New("quota: max_requests must be positive")
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = &quotaState{cfg: cfg, windowAt: time.Now()}
	return nil
}

// Allow reports whether a request to backend is within its quota.
// It increments the counter and resets the window when expired.
func (r *QuotaRegistry) Allow(backend string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.entries[backend]
	if !ok {
		return true // no quota configured — allow
	}
	now := time.Now()
	if now.Sub(s.windowAt) >= s.cfg.Window {
		s.count = 0
		s.windowAt = now
	}
	if s.count >= s.cfg.MaxRequests {
		return false
	}
	s.count++
	return true
}

// Get returns the quota configuration for a backend, if present.
func (r *QuotaRegistry) Get(backend string) (QuotaConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.entries[backend]
	if !ok {
		return QuotaConfig{}, false
	}
	return s.cfg, true
}

// Delete removes the quota for a backend.
func (r *QuotaRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all registered quotas.
func (r *QuotaRegistry) Snapshot() map[string]QuotaConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]QuotaConfig, len(r.entries))
	for k, s := range r.entries {
		out[k] = s.cfg
	}
	return out
}
