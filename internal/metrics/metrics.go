package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot holds a point-in-time copy of the collected metrics.
type Snapshot struct {
	TotalRequests  uint64            `json:"total_requests"`
	ActiveRequests int64             `json:"active_requests"`
	RouteHits      map[string]uint64 `json:"route_hits"`
	UptimeSeconds  float64           `json:"uptime_seconds"`
}

// Collector tracks proxy request metrics.
type Collector struct {
	mu             sync.RWMutex
	totalRequests  uint64
	activeRequests int64
	routeHits      map[string]uint64
	startTime      time.Time
}

// New creates and returns a new Collector.
func New() *Collector {
	return &Collector{
		routeHits: make(map[string]uint64),
		startTime: time.Now(),
	}
}

// IncRequest increments the total and active request counters.
func (c *Collector) IncRequest(route string) {
	atomic.AddUint64(&c.totalRequests, 1)
	atomic.AddInt64(&c.activeRequests, 1)

	c.mu.Lock()
	c.routeHits[route]++
	c.mu.Unlock()
}

// DecActive decrements the active request counter.
func (c *Collector) DecActive() {
	atomic.AddInt64(&c.activeRequests, -1)
}

// Snapshot returns a point-in-time copy of the current metrics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	hits := make(map[string]uint64, len(c.routeHits))
	for k, v := range c.routeHits {
		hits[k] = v
	}
	c.mu.RUnlock()

	return Snapshot{
		TotalRequests:  atomic.LoadUint64(&c.totalRequests),
		ActiveRequests: atomic.LoadInt64(&c.activeRequests),
		RouteHits:      hits,
		UptimeSeconds:  time.Since(c.startTime).Seconds(),
	}
}

// Reset clears all counters (useful for testing).
func (c *Collector) Reset() {
	atomic.StoreUint64(&c.totalRequests, 0)
	atomic.StoreInt64(&c.activeRequests, 0)
	c.mu.Lock()
	c.routeHits = make(map[string]uint64)
	c.mu.Unlock()
	c.startTime = time.Now()
}
