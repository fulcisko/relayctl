package upstream

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrDrained is returned when a backend is in drain mode and no longer accepts new connections.
var ErrDrained = errors.New("upstream: backend is draining")

// DrainRegistry tracks which backends are in drain mode, refusing new traffic
// while allowing in-flight requests to complete.
type DrainRegistry struct {
	mu      sync.RWMutex
	drained map[string]*drainEntry
}

type drainEntry struct {
	active int64 // atomic counter of in-flight requests
}

// NewDrainRegistry creates an empty DrainRegistry.
func NewDrainRegistry() *DrainRegistry {
	return &DrainRegistry{
		drained: make(map[string]*drainEntry),
	}
}

// Drain marks a backend as draining. New requests will be rejected.
func (r *DrainRegistry) Drain(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drained[backend]; !ok {
		r.drained[backend] = &drainEntry{}
	}
}

// Restore removes a backend from drain mode, allowing new traffic.
func (r *DrainRegistry) Restore(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.drained, backend)
}

// IsDraining reports whether the given backend is currently draining.
func (r *DrainRegistry) IsDraining(backend string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.drained[backend]
	return ok
}

// Acquire attempts to acquire a request slot for backend.
// Returns ErrDrained if the backend is draining.
func (r *DrainRegistry) Acquire(backend string) error {
	r.mu.RLock()
	entry, ok := r.drained[backend]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	// Still count in-flight even while draining so callers can track quiescence.
	atomic.AddInt64(&entry.active, 1)
	return ErrDrained
}

// Release decrements the in-flight counter for a draining backend.
func (r *DrainRegistry) Release(backend string) {
	r.mu.RLock()
	entry, ok := r.drained[backend]
	r.mu.RUnlock()
	if ok {
		atomic.AddInt64(&entry.active, -1)
	}
}

// ActiveCount returns the number of in-flight requests for a draining backend.
// Returns 0 if the backend is not registered.
func (r *DrainRegistry) ActiveCount(backend string) int64 {
	r.mu.RLock()
	entry, ok := r.drained[backend]
	r.mu.RUnlock()
	if !ok {
		return 0
	}
	return atomic.LoadInt64(&entry.active)
}

// Snapshot returns a map of backend → active-request-count for all draining backends.
func (r *DrainRegistry) Snapshot() map[string]int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]int64, len(r.drained))
	for k, e := range r.drained {
		out[k] = atomic.LoadInt64(&e.active)
	}
	return out
}
