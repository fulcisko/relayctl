// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"errors"
	"net/http"
	"sync"
)

// ErrNoHealthyBackend is returned when all backends are unavailable.
var ErrNoHealthyBackend = errors.New("upstream: no healthy backend available")

// HealthChecker reports whether a backend URL is considered healthy.
type HealthChecker interface {
	IsHealthy(url string) bool
}

// FailoverBalancer tries backends in priority order, skipping unhealthy ones.
type FailoverBalancer struct {
	mu       sync.RWMutex
	backends []string
	hc       HealthChecker
}

// NewFailoverBalancer creates a FailoverBalancer with an ordered list of backends.
// The first backend is the primary; subsequent entries are fallbacks.
func NewFailoverBalancer(backends []string, hc HealthChecker) (*FailoverBalancer, error) {
	if len(backends) == 0 {
		return nil, errors.New("upstream: failover requires at least one backend")
	}
	return &FailoverBalancer{backends: backends, hc: hc}, nil
}

// Next returns the first healthy backend in priority order.
func (f *FailoverBalancer) Next(_ *http.Request) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, b := range f.backends {
		if f.hc.IsHealthy(b) {
			return b, nil
		}
	}
	return "", ErrNoHealthyBackend
}

// Backends returns a copy of the current backend list.
func (f *FailoverBalancer) Backends() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]string, len(f.backends))
	copy(out, f.backends)
	return out
}

// Update replaces the backend list atomically.
func (f *FailoverBalancer) Update(backends []string) error {
	if len(backends) == 0 {
		return errors.New("upstream: failover requires at least one backend")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.backends = backends
	return nil
}
