package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Status represents the health state of a backend.
type Status struct {
	URL     string
	Healthy bool
	LastChecked time.Time
}

// Checker periodically probes backend URLs and tracks their health.
type Checker struct {
	mu       sync.RWMutex
	statuses map[string]*Status
	interval time.Duration
	timeout  time.Duration
	client   *http.Client
	stopCh   chan struct{}
}

// New creates a new Checker with the given probe interval and timeout.
func New(interval, timeout time.Duration) *Checker {
	return &Checker{
		statuses: make(map[string]*Status),
		interval: interval,
		timeout:  timeout,
		client:   &http.Client{Timeout: timeout},
		stopCh:   make(chan struct{}),
	}
}

// Register adds a backend URL to the set of monitored targets.
func (c *Checker) Register(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.statuses[url]; !exists {
		c.statuses[url] = &Status{URL: url, Healthy: true}
	}
}

// IsHealthy returns whether the given backend URL is currently healthy.
func (c *Checker) IsHealthy(url string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.statuses[url]
	if !ok {
		return true // unknown backends assumed healthy
	}
	return s.Healthy
}

// Snapshot returns a copy of all current health statuses.
func (c *Checker) Snapshot() []Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Status, 0, len(c.statuses))
	for _, s := range c.statuses {
		out = append(out, *s)
	}
	return out
}

// Start begins background health probing. It is non-blocking.
func (c *Checker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.probeAll()
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop halts background probing.
func (c *Checker) Stop() {
	close(c.stopCh)
}

func (c *Checker) probeAll() {
	c.mu.RLock()
	urls := make([]string, 0, len(c.statuses))
	for u := range c.statuses {
		urls = append(urls, u)
	}
	c.mu.RUnlock()

	for _, u := range urls {
		healthy := c.probe(u)
		c.mu.Lock()
		if s, ok := c.statuses[u]; ok {
			s.Healthy = healthy
			s.LastChecked = time.Now()
		}
		c.mu.Unlock()
	}
}

func (c *Checker) probe(url string) bool {
	resp, err := c.client.Get(url)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode < 500
}
