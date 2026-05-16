package upstream

import (
	"errors"
	"sync"
)

// TagRegistry stores arbitrary string tags associated with backends.
// Tags can be used for routing decisions, metadata, or filtering.
type TagRegistry struct {
	mu   sync.RWMutex
	tags map[string][]string
}

var errEmptyTagBackend = errors.New("backend must not be empty")

// NewTagRegistry returns an empty TagRegistry.
func NewTagRegistry() *TagRegistry {
	return &TagRegistry{
		tags: make(map[string][]string),
	}
}

// Set assigns tags to a backend, replacing any existing tags.
func (r *TagRegistry) Set(backend string, tags []string) error {
	if backend == "" {
		return errEmptyTagBackend
	}
	copy := make([]string, len(tags))
	for i, t := range tags {
		copy[i] = t
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[backend] = copy
	return nil
}

// Get returns the tags for a backend and whether the backend was found.
func (r *TagRegistry) Get(backend string) ([]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tags[backend]
	if !ok {
		return nil, false
	}
	copy := make([]string, len(t))
	for i, v := range t {
		copy[i] = v
	}
	return copy, true
}

// Delete removes all tags for a backend.
func (r *TagRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tags, backend)
}

// Snapshot returns a copy of all backend-to-tags mappings.
func (r *TagRegistry) Snapshot() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]string, len(r.tags))
	for k, v := range r.tags {
		copy := make([]string, len(v))
		for i, t := range v {
			copy[i] = t
		}
		out[k] = copy
	}
	return out
}
