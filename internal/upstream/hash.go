// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
)

// HashBalancer selects a backend deterministically based on a request attribute
// (e.g. client IP or a header value), providing session affinity without cookies.
type HashBalancer struct {
	mu       sync.RWMutex
	backends []string
	keyFn    func(r *http.Request) string
}

// NewHashBalancer creates a HashBalancer using keyFn to derive a hash key from
// each request. Returns an error if backends is empty or keyFn is nil.
func NewHashBalancer(backends []string, keyFn func(r *http.Request) string) (*HashBalancer, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("hash balancer: at least one backend required")
	}
	if keyFn == nil {
		return nil, fmt.Errorf("hash balancer: keyFn must not be nil")
	}
	cp := make([]string, len(backends))
	copy(cp, backends)
	return &HashBalancer{backends: cp, keyFn: keyFn}, nil
}

// Next returns the backend selected by hashing the key derived from r.
func (h *HashBalancer) Next(r *http.Request) (string, error) {
	key := h.keyFn(r)

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))

	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.backends) == 0 {
		return "", fmt.Errorf("hash balancer: no backends available")
	}

	idx := int(hash.Sum32()) % len(h.backends)
	return h.backends[idx], nil
}

// Backends returns a copy of the current backend list.
func (h *HashBalancer) Backends() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	cp := make([]string, len(h.backends))
	copy(cp, h.backends)
	return cp
}

// Update replaces the backend list. Returns an error if the new list is empty.
func (h *HashBalancer) Update(backends []string) error {
	if len(backends) == 0 {
		return fmt.Errorf("hash balancer: at least one backend required")
	}
	cp := make([]string, len(backends))
	copy(cp, backends)
	h.mu.Lock()
	h.backends = cp
	h.mu.Unlock()
	return nil
}

// IPKeyFn is a convenience keyFn that uses the request's remote IP as the hash key.
func IPKeyFn(r *http.Request) string {
	return ipFromAddr(r.RemoteAddr)
}

// HeaderKeyFn returns a keyFn that hashes the value of the named HTTP header.
func HeaderKeyFn(header string) func(r *http.Request) string {
	return func(r *http.Request) string {
		v := r.Header.Get(header)
		if v == "" {
			return r.RemoteAddr
		}
		return v
	}
}

// ipFromAddr strips the port from a host:port string.
func ipFromAddr(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
