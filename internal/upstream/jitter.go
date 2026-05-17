package upstream

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// jitterEntry holds the base and max jitter duration for a backend.
type jitterEntry struct {
	base   time.Duration
	jitter time.Duration
}

// JitterRegistry stores per-backend jitter configuration.
type JitterRegistry struct {
	mu      sync.RWMutex
	entries map[string]jitterEntry
	rng     *rand.Rand
}

// NewJitterRegistry returns an empty JitterRegistry.
func NewJitterRegistry() *JitterRegistry {
	return &JitterRegistry{
		entries: make(map[string]jitterEntry),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Set registers a jitter policy for the given backend.
// base is the minimum delay; jitter is the maximum additional random delay.
// Returns an error if backend is empty or either duration is negative.
func (r *JitterRegistry) Set(backend string, base, jitter time.Duration) error {
	if backend == "" {
		return errors.New("jitter: backend must not be empty")
	}
	if base < 0 {
		return errors.New("jitter: base duration must not be negative")
	}
	if jitter < 0 {
		return errors.New("jitter: jitter duration must not be negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = jitterEntry{base: base, jitter: jitter}
	return nil
}

// Delay returns the computed delay for the given backend.
// If no entry exists, it returns zero.
func (r *JitterRegistry) Delay(backend string) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[backend]
	if !ok {
		return 0
	}
	if e.jitter == 0 {
		return e.base
	}
	return e.base + time.Duration(r.rng.Int63n(int64(e.jitter)))
}

// Get returns the base and jitter configuration for a backend.
// ok is false if no entry is registered.
func (r *JitterRegistry) Get(backend string) (base, jitter time.Duration, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[backend]
	return e.base, e.jitter, ok
}

// Delete removes the jitter policy for the given backend.
func (r *JitterRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all registered entries as a map of
// backend -> [base, jitter].
func (r *JitterRegistry) Snapshot() map[string][2]time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][2]time.Duration, len(r.entries))
	for k, v := range r.entries {
		out[k] = [2]time.Duration{v.base, v.jitter}
	}
	return out
}
