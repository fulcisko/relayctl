package upstream

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// BackoffPolicy defines exponential backoff parameters for a backend.
type BackoffPolicy struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     float64
}

type backoffEntry struct {
	policy  BackoffPolicy
	attempt int
	mu      sync.Mutex
}

func (e *backoffEntry) next() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	delay := float64(e.policy.BaseDelay) * math.Pow(e.policy.Multiplier, float64(e.attempt))
	if delay > float64(e.policy.MaxDelay) {
		delay = float64(e.policy.MaxDelay)
	}
	if e.policy.Jitter > 0 {
		delay += delay * e.policy.Jitter * (rand.Float64()*2 - 1)
	}
	e.attempt++
	return time.Duration(delay)
}

func (e *backoffEntry) reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.attempt = 0
}

// BackoffRegistry stores per-backend exponential backoff policies.
type BackoffRegistry struct {
	mu      sync.RWMutex
	entries map[string]*backoffEntry
}

// NewBackoffRegistry returns an empty BackoffRegistry.
func NewBackoffRegistry() *BackoffRegistry {
	return &BackoffRegistry{entries: make(map[string]*backoffEntry)}
}

// Set registers or replaces the backoff policy for a backend.
func (r *BackoffRegistry) Set(backend string, p BackoffPolicy) error {
	if backend == "" {
		return errEmptyBackend
	}
	if p.Multiplier <= 0 {
		p.Multiplier = 2.0
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = 30 * time.Second
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = &backoffEntry{policy: p}
	return nil
}

// Next returns the next backoff duration for the backend, advancing its attempt counter.
func (r *BackoffRegistry) Next(backend string) (time.Duration, bool) {
	r.mu.RLock()
	e, ok := r.entries[backend]
	r.mu.RUnlock()
	if !ok {
		return 0, false
	}
	return e.next(), true
}

// Reset resets the attempt counter for the backend.
func (r *BackoffRegistry) Reset(backend string) {
	r.mu.RLock()
	e, ok := r.entries[backend]
	r.mu.RUnlock()
	if ok {
		e.reset()
	}
}

// Delete removes the backoff policy for a backend.
func (r *BackoffRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a map of backend to policy for all registered entries.
func (r *BackoffRegistry) Snapshot() map[string]BackoffPolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]BackoffPolicy, len(r.entries))
	for k, e := range r.entries {
		out[k] = e.policy
	}
	return out
}
