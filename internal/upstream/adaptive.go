// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"sync"
	"time"
)

// AdaptiveBalancer selects backends based on their recent response latency,
// preferring backends with lower average response times.
type AdaptiveBalancer struct {
	mu       sync.Mutex
	backends []string
	latency  map[string]time.Duration
	samples  map[string]int
}

// NewAdaptiveBalancer creates an AdaptiveBalancer from the given backend list.
// Returns an error if the list is empty.
func NewAdaptiveBalancer(backends []string) (*AdaptiveBalancer, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}
	b := &AdaptiveBalancer{
		backends: make([]string, len(backends)),
		latency:  make(map[string]time.Duration),
		samples:  make(map[string]int),
	}
	copy(b.backends, backends)
	for _, addr := range backends {
		b.latency[addr] = 0
		b.samples[addr] = 0
	}
	return b, nil
}

// Next returns the backend with the lowest recorded average latency.
// Backends with no samples are treated as having zero latency (preferred).
func (b *AdaptiveBalancer) Next(_ string) (string, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	best := b.backends[0]
	bestLatency := b.latency[best]

	for _, addr := range b.backends[1:] {
		if b.latency[addr] < bestLatency {
			best = addr
			bestLatency = b.latency[addr]
		}
	}

	start := time.Now()
	done := func() {
		elapsed := time.Since(start)
		b.mu.Lock()
		defer b.mu.Unlock()
		n := b.samples[best]
		// Exponential moving average with weight 0.2 for new sample.
		if n == 0 {
			b.latency[best] = elapsed
		} else {
			b.latency[best] = time.Duration(float64(b.latency[best])*0.8 + float64(elapsed)*0.2)
		}
		b.samples[best] = n + 1
	}
	return best, done
}

// Backends returns a copy of the backend list.
func (b *AdaptiveBalancer) Backends() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.backends))
	copy(out, b.backends)
	return out
}

// AvgLatency returns the recorded average latency for a given backend.
func (b *AdaptiveBalancer) AvgLatency(addr string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.latency[addr]
}
