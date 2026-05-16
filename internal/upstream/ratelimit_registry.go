package upstream

import (
	"errors"
	"sync"
)

// PerBackendRateLimit holds rate limit settings for a single backend.
type PerBackendRateLimit struct {
	// RequestsPerSecond is the maximum allowed requests per second.
	RequestsPerSecond float64
	// Burst is the maximum burst size.
	Burst int
}

// RateLimitRegistry stores per-backend rate limit configurations.
type RateLimitRegistry struct {
	mu      sync.RWMutex
	entries map[string]PerBackendRateLimit
}

// NewRateLimitRegistry returns an empty RateLimitRegistry.
func NewRateLimitRegistry() *RateLimitRegistry {
	return &RateLimitRegistry{
		entries: make(map[string]PerBackendRateLimit),
	}
}

// Set stores a rate limit config for the given backend.
func (r *RateLimitRegistry) Set(backend string, cfg PerBackendRateLimit) error {
	if backend == "" {
		return errors.New("backend must not be empty")
	}
	if cfg.RequestsPerSecond <= 0 {
		return errors.New("requests_per_second must be > 0")
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 1
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = cfg
	return nil
}

// Get returns the rate limit config for the given backend.
func (r *RateLimitRegistry) Get(backend string) (PerBackendRateLimit, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.entries[backend]
	return cfg, ok
}

// Delete removes the rate limit config for the given backend.
func (r *RateLimitRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all current entries.
func (r *RateLimitRegistry) Snapshot() map[string]PerBackendRateLimit {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]PerBackendRateLimit, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}
