// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"errors"
	"sync"
)

// WeightedBackend pairs a backend URL with its relative weight.
type WeightedBackend struct {
	URL    string
	Weight int
}

// WeightedBalancer selects backends using weighted round-robin.
type WeightedBalancer struct {
	mu       sync.Mutex
	backends []WeightedBackend
	current  int
	counter  int
}

// NewWeightedBalancer creates a WeightedBalancer from the given backends.
// All weights must be positive; returns an error if the slice is empty or
// any weight is <= 0.
func NewWeightedBalancer(backends []WeightedBackend) (*WeightedBalancer, error) {
	if len(backends) == 0 {
		return nil, errors.New("weighted balancer: no backends provided")
	}
	for _, b := range backends {
		if b.Weight <= 0 {
			return nil, errors.New("weighted balancer: all weights must be positive")
		}
	}
	cp := make([]WeightedBackend, len(backends))
	copy(cp, backends)
	return &WeightedBalancer{backends: cp}, nil
}

// Next returns the next backend URL according to weighted round-robin.
func (w *WeightedBalancer) Next() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	for {
		b := w.backends[w.current]
		if w.counter < b.Weight {
			w.counter++
			return b.URL
		}
		w.counter = 0
		w.current = (w.current + 1) % len(w.backends)
	}
}

// Backends returns a snapshot of the current weighted backends.
func (w *WeightedBalancer) Backends() []WeightedBackend {
	w.mu.Lock()
	defer w.mu.Unlock()
	cp := make([]WeightedBackend, len(w.backends))
	copy(cp, w.backends)
	return cp
}
