package upstream

import "sync"

// PassthroughRegistry tracks backends that should bypass middleware processing.
type PassthroughRegistry struct {
	mu      sync.RWMutex
	entries map[string]struct{}
}

// NewPassthroughRegistry creates an empty PassthroughRegistry.
func NewPassthroughRegistry() *PassthroughRegistry {
	return &PassthroughRegistry{
		entries: make(map[string]struct{}),
	}
}

// Set marks the given backend URL as a passthrough target.
// Returns an error if backend is empty.
func (r *PassthroughRegistry) Set(backend string) error {
	if backend == "" {
		return errEmptyBackend
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = struct{}{}
	return nil
}

// Delete removes the passthrough marking for the given backend.
func (r *PassthroughRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// IsPassthrough reports whether the backend is registered as a passthrough target.
func (r *PassthroughRegistry) IsPassthrough(backend string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.entries[backend]
	return ok
}

// Snapshot returns a copy of all registered passthrough backends.
func (r *PassthroughRegistry) Snapshot() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.entries))
	for k := range r.entries {
		out = append(out, k)
	}
	return out
}
