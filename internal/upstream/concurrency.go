package upstream

import (
	"errors"
	"fmt"
	"sync"
)

// ErrConcurrencyLimitExceeded is returned when a backend is at capacity.
var ErrConcurrencyLimitExceeded = errors.New("concurrency limit exceeded")

// ConcurrencyToken must be released after the request completes.
type ConcurrencyToken struct {
	release func()
}

// Release decrements the in-flight counter for the backend.
func (t *ConcurrencyToken) Release() { t.release() }

type concurrencyEntry struct {
	max    int64
	active int64
	mu     sync.Mutex
}

// ConcurrencyRegistry tracks per-backend concurrency limits.
type ConcurrencyRegistry struct {
	mu      sync.RWMutex
	entries map[string]*concurrencyEntry
}

// NewConcurrencyRegistry returns an empty ConcurrencyRegistry.
func NewConcurrencyRegistry() *ConcurrencyRegistry {
	return &ConcurrencyRegistry{entries: make(map[string]*concurrencyEntry)}
}

// Set configures a maximum concurrency for the given backend.
func (r *ConcurrencyRegistry) Set(backend string, max int64) error {
	if backend == "" {
		return errors.New("backend must not be empty")
	}
	if max <= 0 {
		return fmt.Errorf("max must be > 0, got %d", max)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = &concurrencyEntry{max: max}
	return nil
}

// Get returns the configured limit and current active count, or false if unset.
func (r *ConcurrencyRegistry) Get(backend string) (max, active int64, ok bool) {
	r.mu.RLock()
	e, exists := r.entries[backend]
	r.mu.RUnlock()
	if !exists {
		return 0, 0, false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.max, e.active, true
}

// Acquire attempts to reserve a slot for the backend.
// Returns ErrConcurrencyLimitExceeded if the backend is at capacity.
// If no limit is configured, it returns a no-op token.
func (r *ConcurrencyRegistry) Acquire(backend string) (*ConcurrencyToken, error) {
	r.mu.RLock()
	e, exists := r.entries[backend]
	r.mu.RUnlock()
	if !exists {
		return &ConcurrencyToken{release: func() {}}, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.active >= e.max {
		return nil, ErrConcurrencyLimitExceeded
	}
	e.active++
	return &ConcurrencyToken{release: func() {
		e.mu.Lock()
		e.active--
		e.mu.Unlock()
	}}, nil
}

// Delete removes the concurrency limit for a backend.
func (r *ConcurrencyRegistry) Delete(backend string) {
	r.mu.Lock()
	delete(r.entries, backend)
	r.mu.Unlock()
}
