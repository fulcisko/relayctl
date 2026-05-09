package circuitbreaker

import (
	"sync"
)

// Registry manages a collection of CircuitBreakers keyed by backend URL.
type Registry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   Config
}

// Config holds default circuit breaker settings for the registry.
type Config struct {
	MaxFailures uint
	OpenTimeout  int // seconds
}

// NewRegistry creates a new Registry with the given default config.
func NewRegistry(cfg Config) *Registry {
	return &Registry{
		breakers: make(map[string]*CircuitBreaker),
		config:   cfg,
	}
}

// Get returns the CircuitBreaker for the given key, creating one if needed.
func (r *Registry) Get(key string) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[key]
	r.mu.RUnlock()
	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	// Double-check after acquiring write lock.
	if cb, ok = r.breakers[key]; ok {
		return cb
	}
	cb = New(r.config.MaxFailures, r.config.OpenTimeout)
	r.breakers[key] = cb
	return cb
}

// Reset resets the CircuitBreaker for the given key.
func (r *Registry) Reset(key string) {
	r.mu.RLock()
	cb, ok := r.breakers[key]
	r.mu.RUnlock()
	if ok {
		cb.Reset()
	}
}

// Snapshot returns a map of key -> State for all tracked breakers.
func (r *Registry) Snapshot() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.breakers))
	for k, cb := range r.breakers {
		out[k] = cb.State()
	}
	return out
}

// Keys returns all registered backend keys.
func (r *Registry) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.breakers))
	for k := range r.breakers {
		keys = append(keys, k)
	}
	return keys
}
