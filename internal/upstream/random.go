// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"errors"
	"math/rand"
	"sync"
)

// RandomBalancer selects a backend at random on each request.
type RandomBalancer struct {
	mu       sync.RWMutex
	backends []string
	rng      *rand.Rand
}

// NewRandomBalancer creates a RandomBalancer from the given backend list.
// Returns an error if the list is empty.
func NewRandomBalancer(backends []string, src rand.Source) (*RandomBalancer, error) {
	if len(backends) == 0 {
		return nil, errors.New("random balancer: at least one backend required")
	}
	if src == nil {
		src = rand.NewSource(rand.Int63())
	}
	cp := make([]string, len(backends))
	copy(cp, backends)
	return &RandomBalancer{
		backends: cp,
		rng:      rand.New(src),
	}, nil
}

// Next returns a randomly selected backend URL.
func (r *RandomBalancer) Next(_ string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.backends) == 0 {
		return "", errors.New("random balancer: no backends available")
	}
	return r.backends[r.rng.Intn(len(r.backends))], nil
}

// Backends returns a copy of the current backend list.
func (r *RandomBalancer) Backends() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]string, len(r.backends))
	copy(cp, r.backends)
	return cp
}

// Update replaces the backend list atomically.
func (r *RandomBalancer) Update(backends []string) error {
	if len(backends) == 0 {
		return errors.New("random balancer: at least one backend required")
	}
	cp := make([]string, len(backends))
	copy(cp, backends)
	r.mu.Lock()
	r.backends = cp
	r.mu.Unlock()
	return nil
}
