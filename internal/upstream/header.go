// Package upstream provides load balancing and routing primitives.
package upstream

import (
	"net/http"
	"sync"
)

// HeaderRule defines a header mutation to apply to proxied requests.
type HeaderRule struct {
	Set    map[string]string `json:"set,omitempty"`
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

// HeaderRegistry stores per-backend header mutation rules.
type HeaderRegistry struct {
	mu    sync.RWMutex
	rules map[string]HeaderRule
}

// NewHeaderRegistry creates an empty HeaderRegistry.
func NewHeaderRegistry() *HeaderRegistry {
	return &HeaderRegistry{rules: make(map[string]HeaderRule)}
}

// Set stores a HeaderRule for the given backend URL.
func (r *HeaderRegistry) Set(backend string, rule HeaderRule) error {
	if backend == "" {
		return errEmptyBackend
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[backend] = rule
	return nil
}

// Get retrieves the HeaderRule for the given backend, and whether it exists.
func (r *HeaderRegistry) Get(backend string) (HeaderRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[backend]
	return rule, ok
}

// Delete removes the HeaderRule for the given backend.
func (r *HeaderRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules, backend)
}

// Snapshot returns a copy of all registered rules.
func (r *HeaderRegistry) Snapshot() map[string]HeaderRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]HeaderRule, len(r.rules))
	for k, v := range r.rules {
		out[k] = v
	}
	return out
}

// Apply mutates req headers according to the rule registered for backend.
// If no rule exists the request is left unchanged.
func (r *HeaderRegistry) Apply(backend string, req *http.Request) {
	rule, ok := r.Get(backend)
	if !ok {
		return
	}
	for k, v := range rule.Set {
		req.Header.Set(k, v)
	}
	for k, v := range rule.Add {
		req.Header.Add(k, v)
	}
	for _, k := range rule.Remove {
		req.Header.Del(k)
	}
}
