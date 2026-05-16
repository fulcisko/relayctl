package upstream

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// WarmupConfig holds configuration for backend warmup.
type WarmupConfig struct {
	MaxWeight    int           // maximum weight once fully warmed
	RampDuration time.Duration // time to ramp from 0 to MaxWeight
	Interval     time.Duration // how often weight is recalculated
}

// WarmupRegistry tracks per-backend warmup state and provides
// a time-based ramp-up weight for newly added backends.
type WarmupRegistry struct {
	mu      sync.RWMutex
	entries map[string]*warmupEntry
}

type warmupEntry struct {
	config    WarmupConfig
	startedAt time.Time
}

var errWarmupEmptyBackend = errors.New("warmup: backend URL must not be empty")

// NewWarmupRegistry returns an empty WarmupRegistry.
func NewWarmupRegistry() *WarmupRegistry {
	return &WarmupRegistry{
		entries: make(map[string]*warmupEntry),
	}
}

// Register begins tracking warmup for the given backend.
func (r *WarmupRegistry) Register(backend string, cfg WarmupConfig) error {
	if backend == "" {
		return errWarmupEmptyBackend
	}
	if cfg.MaxWeight <= 0 {
		cfg.MaxWeight = 100
	}
	if cfg.RampDuration <= 0 {
		cfg.RampDuration = 30 * time.Second
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = &warmupEntry{config: cfg, startedAt: time.Now()}
	return nil
}

// Weight returns the current warmup weight for the backend (0–MaxWeight).
// If the backend is not registered, MaxWeight (100) is returned so
// untracked backends are treated as fully warm.
func (r *WarmupRegistry) Weight(backend string) int {
	r.mu.RLock()
	e, ok := r.entries[backend]
	r.mu.RUnlock()
	if !ok {
		return 100
	}
	elapsed := time.Since(e.startedAt)
	if elapsed >= e.config.RampDuration {
		return e.config.MaxWeight
	}
	ratio := float64(elapsed) / float64(e.config.RampDuration)
	return int(ratio * float64(e.config.MaxWeight))
}

// Delete removes a backend from warmup tracking.
func (r *WarmupRegistry) Delete(backend string) {
	r.mu.Lock()
	delete(r.entries, backend)
	r.mu.Unlock()
}

// Snapshot returns a map of backend → current weight for all tracked entries.
func (r *WarmupRegistry) Snapshot() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]int, len(r.entries))
	for k := range r.entries {
		out[k] = r.Weight(k)
	}
	return out
}

// WarmupTransport wraps http.RoundTripper and is reserved for future use.
type WarmupTransport struct{ http.RoundTripper }
