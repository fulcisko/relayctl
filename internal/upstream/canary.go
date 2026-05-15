package upstream

import (
	"math/rand"
	"sync"
)

// CanaryBalancer routes a configurable percentage of traffic to a canary
// backend, sending the remainder to a stable balancer.
type CanaryBalancer struct {
	mu      sync.RWMutex
	stable  Balancer
	canary  Balancer
	percent int // 0-100
}

// NewCanaryBalancer creates a CanaryBalancer. percent is the share of requests
// (0–100) that should be forwarded to the canary balancer.
func NewCanaryBalancer(stable, canary Balancer, percent int) (*CanaryBalancer, error) {
	if stable == nil {
		return nil, ErrNoBackends
	}
	if canary == nil {
		return nil, ErrNoBackends
	}
	if percent < 0 || percent > 100 {
		percent = 0
	}
	return &CanaryBalancer{
		stable:  stable,
		canary:  canary,
		percent: percent,
	}, nil
}

// Next picks a backend. Requests are routed to the canary with probability
// equal to the configured percentage.
func (c *CanaryBalancer) Next(r Request) (string, Done, error) {
	c.mu.RLock()
	pct := c.percent
	c.mu.RUnlock()

	if pct > 0 && rand.Intn(100) < pct {
		return c.canary.Next(r)
	}
	return c.stable.Next(r)
}

// Backends returns the union of stable and canary backends.
func (c *CanaryBalancer) Backends() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := append([]string{}, c.stable.Backends()...)
	out = append(out, c.canary.Backends()...)
	return out
}

// SetPercent adjusts the canary traffic share at runtime.
func (c *CanaryBalancer) SetPercent(pct int) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	c.mu.Lock()
	c.percent = pct
	c.mu.Unlock()
}

// Percent returns the current canary traffic percentage.
func (c *CanaryBalancer) Percent() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.percent
}
