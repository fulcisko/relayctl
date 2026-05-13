package upstream

import (
	"errors"
	"sync/atomic"
)

// ErrNoBackends is returned when the backend pool is empty.
var ErrNoBackends = errors.New("upstream: no backends available")

// Balancer selects a backend from a pool using round-robin.
type Balancer struct {
	backends []string
	counter  atomic.Uint64
}

// New creates a Balancer with the given backend URLs.
func New(backends []string) (*Balancer, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}
	return &Balancer{backends: backends}, nil
}

// Next returns the next backend URL in round-robin order.
func (b *Balancer) Next() (string, error) {
	if len(b.backends) == 0 {
		return "", ErrNoBackends
	}
	idx := b.counter.Add(1) - 1
	return b.backends[idx%uint64(len(b.backends))], nil
}

// Backends returns a copy of the current backend list.
func (b *Balancer) Backends() []string {
	result := make([]string, len(b.backends))
	copy(result, b.backends)
	return result
}

// Update replaces the backend pool atomically.
func (b *Balancer) Update(backends []string) error {
	if len(backends) == 0 {
		return ErrNoBackends
	}
	b.backends = backends
	b.counter.Store(0)
	return nil
}
