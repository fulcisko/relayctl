package upstream

import (
	"net/http"
	"sync"
)

// ResponseHeaderRegistry stores per-backend response header mutation rules.
// Unlike HeaderRegistry (which operates on requests), this registry is applied
// after the upstream response is received, before it is written to the client.
type ResponseHeaderRegistry struct {
	mu    sync.RWMutex
	rules map[string]ResponseHeaderRules
}

// ResponseHeaderRules defines headers to set or delete on an upstream response.
type ResponseHeaderRules struct {
	Set map[string]string `json:"set,omitempty"`
	Del []string          `json:"del,omitempty"`
}

// NewResponseHeaderRegistry returns an initialised ResponseHeaderRegistry.
func NewResponseHeaderRegistry() *ResponseHeaderRegistry {
	return &ResponseHeaderRegistry{
		rules: make(map[string]ResponseHeaderRules),
	}
}

// Set stores rules for the given backend URL, replacing any existing entry.
func (r *ResponseHeaderRegistry) Set(backend string, rules ResponseHeaderRules) error {
	if backend == "" {
		return errEmptyBackend
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[backend] = rules
	return nil
}

// Get returns the rules for the given backend and whether they were found.
func (r *ResponseHeaderRegistry) Get(backend string) (ResponseHeaderRules, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.rules[backend]
	return v, ok
}

// Delete removes the rules for the given backend.
func (r *ResponseHeaderRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules, backend)
}

// Apply mutates resp according to any rules registered for backend.
// It is a no-op when no rules exist for the backend.
func (r *ResponseHeaderRegistry) Apply(backend string, resp *http.Response) {
	if resp == nil {
		return
	}
	r.mu.RLock()
	rules, ok := r.rules[backend]
	r.mu.RUnlock()
	if !ok {
		return
	}
	for k, v := range rules.Set {
		resp.Header.Set(k, v)
	}
	for _, k := range rules.Del {
		resp.Header.Del(k)
	}
}

// Snapshot returns a copy of all registered rules keyed by backend URL.
func (r *ResponseHeaderRegistry) Snapshot() map[string]ResponseHeaderRules {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]ResponseHeaderRules, len(r.rules))
	for k, v := range r.rules {
		out[k] = v
	}
	return out
}
