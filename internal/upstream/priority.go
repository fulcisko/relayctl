package upstream

import (
	"errors"
	"net/http"
	"sync"
)

// PriorityBalancer routes requests to the highest-priority healthy backend.
// Backends are tried in descending priority order; the first healthy one wins.
// If all backends are unhealthy the lowest-priority backend is returned as a
// last-resort fallback.
type PriorityBalancer struct {
	mu       sync.RWMutex
	backends []priorityEntry
	hc       HealthChecker
}

type priorityEntry struct {
	backend  string
	priority int
}

// HealthChecker is the subset of healthcheck.Checker used by PriorityBalancer.
type HealthChecker interface {
	IsHealthy(url string) bool
}

var errNoPriorityBackends = errors.New("priority: no backends configured")

// NewPriorityBalancer creates a PriorityBalancer from a map of backend→priority.
// Higher integer values mean higher priority.
func NewPriorityBalancer(backends map[string]int, hc HealthChecker) (*PriorityBalancer, error) {
	if len(backends) == 0 {
		return nil, errNoPriorityBackends
	}
	if hc == nil {
		return nil, errors.New("priority: health checker must not be nil")
	}
	entries := make([]priorityEntry, 0, len(backends))
	for b, p := range backends {
		entries = append(entries, priorityEntry{backend: b, priority: p})
	}
	// sort descending by priority
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].priority > entries[j-1].priority; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
	return &PriorityBalancer{backends: entries, hc: hc}, nil
}

// Next returns the highest-priority healthy backend URL.
func (p *PriorityBalancer) Next(r *http.Request) string {
	p.mu.RLock()
	entries := p.backends
	p.mu.RUnlock()
	for _, e := range entries {
		if p.hc.IsHealthy(e.backend) {
			return e.backend
		}
	}
	if len(entries) > 0 {
		return entries[len(entries)-1].backend
	}
	return ""
}

// Backends returns a copy of the ordered backend list.
func (p *PriorityBalancer) Backends() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]string, len(p.backends))
	for i, e := range p.backends {
		out[i] = e.backend
	}
	return out
}
