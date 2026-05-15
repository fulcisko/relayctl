// Package upstream provides load balancing strategies for backend selection.
// burst.go implements a burst-aware balancer that tracks request spikes
// and temporarily expands the active backend pool during high-traffic windows.
package upstream

import (
	"sync"
	"time"
)

// BurstBalancer wraps a primary balancer and activates burst backends
// when the request rate exceeds a configured threshold within a window.
type BurstBalancer struct {
	mu          sync.Mutex
	primary     Balancer
	burst       Balancer
	threshold   int           // requests per window to trigger burst
	window      time.Duration
	count       int
	windowStart time.Time
	burstActive bool
	burstUntil  time.Time
	burstTTL    time.Duration
}

// NewBurstBalancer creates a BurstBalancer. When request count exceeds
// threshold within window, burst backends are included for burstTTL duration.
func NewBurstBalancer(primary, burst Balancer, threshold int, window, burstTTL time.Duration) (*BurstBalancer, error) {
	if primary == nil {
		return nil, errNilBalancer("primary")
	}
	if burst == nil {
		return nil, errNilBalancer("burst")
	}
	if threshold <= 0 {
		threshold = 100
	}
	if window <= 0 {
		window = time.Second
	}
	if burstTTL <= 0 {
		burstTTL = 5 * time.Second
	}
	return &BurstBalancer{
		primary:     primary,
		burst:       burst,
		threshold:   threshold,
		window:      window,
		windowStart: time.Now(),
		burstTTL:    burstTTL,
	}, nil
}

// Next selects a backend, routing to burst pool when burst mode is active.
func (b *BurstBalancer) Next(r Request) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	if now.Sub(b.windowStart) > b.window {
		b.count = 0
		b.windowStart = now
	}
	b.count++

	if b.count >= b.threshold && !b.burstActive {
		b.burstActive = true
		b.burstUntil = now.Add(b.burstTTL)
	}
	if b.burstActive && now.After(b.burstUntil) {
		b.burstActive = false
	}

	if b.burstActive {
		if backend, err := b.burst.Next(r); err == nil {
			return backend, nil
		}
	}
	return b.primary.Next(r)
}

// Backends returns the union of primary and burst backends.
func (b *BurstBalancer) Backends() []string {
	seen := make(map[string]struct{})
	var all []string
	for _, backend := range b.primary.Backends() {
		if _, ok := seen[backend]; !ok {
			seen[backend] = struct{}{}
			all = append(all, backend)
		}
	}
	for _, backend := range b.burst.Backends() {
		if _, ok := seen[backend]; !ok {
			seen[backend] = struct{}{}
			all = append(all, backend)
		}
	}
	return all
}

// IsBurstActive reports whether burst mode is currently engaged.
func (b *BurstBalancer) IsBurstActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.burstActive && time.Now().Before(b.burstUntil)
}

func errNilBalancer(name string) error {
	return fmt.Errorf("burst: %s balancer must not be nil", name)
}
