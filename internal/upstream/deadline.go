package upstream

import (
	"errors"
	"sync"
	"time"
)

// DeadlineRegistry stores per-backend absolute request deadlines.
type DeadlineRegistry struct {
	mu      sync.RWMutex
	entries map[string]time.Duration
}

// NewDeadlineRegistry returns an empty DeadlineRegistry.
func NewDeadlineRegistry() *DeadlineRegistry {
	return &DeadlineRegistry{
		entries: make(map[string]time.Duration),
	}
}

// Set registers a deadline duration for the given backend.
// Returns an error if backend is empty or duration is non-positive.
func (r *DeadlineRegistry) Set(backend string, d time.Duration) error {
	if backend == "" {
		return errors.New("deadline: backend must not be empty")
	}
	if d <= 0 {
		return errors.New("deadline: duration must be positive")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = d
	return nil
}

// Get returns the deadline duration for the given backend.
// The second return value is false if no deadline is registered.
func (r *DeadlineRegistry) Get(backend string) (time.Duration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.entries[backend]
	return d, ok
}

// Delete removes the deadline for the given backend.
func (r *DeadlineRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all registered deadlines.
func (r *DeadlineRegistry) Snapshot() map[string]time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]time.Duration, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}
