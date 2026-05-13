// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"net/http"
	"sync"
	"sync/atomic"
)

// connBackend tracks active connection count for a backend.
type connBackend struct {
	url   string
	count atomic.Int64
}

// LeastConnBalancer selects the backend with the fewest active connections.
type LeastConnBalancer struct {
	mu       sync.RWMutex
	backends []*connBackend
}

// NewLeastConnBalancer creates a LeastConnBalancer from a list of backend URLs.
// Returns an error if backends is empty.
func NewLeastConnBalancer(urls []string) (*LeastConnBalancer, error) {
	if len(urls) == 0 {
		return nil, ErrNoBackends
	}
	bs := make([]*connBackend, len(urls))
	for i, u := range urls {
		bs[i] = &connBackend{url: u}
	}
	return &LeastConnBalancer{backends: bs}, nil
}

// Next returns the backend URL with the fewest active connections and
// increments its counter. The caller must call Done when the request finishes.
func (lb *LeastConnBalancer) Next(_ *http.Request) (string, func()) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var chosen *connBackend
	for _, b := range lb.backends {
		if chosen == nil || b.count.Load() < chosen.count.Load() {
			chosen = b
		}
	}
	if chosen == nil {
		return "", func() {}
	}
	chosen.count.Add(1)
	return chosen.url, func() { chosen.count.Add(-1) }
}

// Backends returns a snapshot of current backend URLs.
func (lb *LeastConnBalancer) Backends() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	out := make([]string, len(lb.backends))
	for i, b := range lb.backends {
		out[i] = b.url
	}
	return out
}

// ActiveCount returns the current active connection count for a given URL.
// Returns -1 if the URL is not found.
func (lb *LeastConnBalancer) ActiveCount(url string) int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	for _, b := range lb.backends {
		if b.url == url {
			return b.count.Load()
		}
	}
	return -1
}
