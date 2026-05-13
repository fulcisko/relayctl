package upstream

import (
	"sync"
)

// HealthChecker is an interface for checking backend health.
type HealthChecker interface {
	IsHealthy(url string) bool
}

// FilteredBalancer wraps a Balancer and skips unhealthy backends.
type FilteredBalancer struct {
	mu      sync.RWMutex
	bal     *Balancer
	health  HealthChecker
}

// NewFilteredBalancer creates a FilteredBalancer using the given balancer and health checker.
func NewFilteredBalancer(bal *Balancer, hc HealthChecker) *FilteredBalancer {
	return &FilteredBalancer{bal: bal, health: hc}
}

// Next returns the next healthy backend URL, skipping unhealthy ones.
// Returns an empty string if no healthy backend is available.
func (fb *FilteredBalancer) Next() string {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	backends := fb.bal.Backends()
	n := len(backends)
	if n == 0 {
		return ""
	}
	for i := 0; i < n; i++ {
		candidate := fb.bal.Next()
		if fb.health.IsHealthy(candidate) {
			return candidate
		}
	}
	return ""
}

// UpdateBalancer replaces the underlying balancer.
func (fb *FilteredBalancer) UpdateBalancer(bal *Balancer) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	fb.bal = bal
}

// HealthyBackends returns the list of currently healthy backends.
func (fb *FilteredBalancer) HealthyBackends() []string {
	fb.mu.RLock()
	defer fb.mu.RUnlock()

	var healthy []string
	for _, b := range fb.bal.Backends() {
		if fb.health.IsHealthy(b) {
			healthy = append(healthy, b)
		}
	}
	return healthy
}
