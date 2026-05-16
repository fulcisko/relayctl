package upstream

import (
	"net/http"
	"sync"
	"time"
)

// CacheEntry holds a cached response body and metadata.
type CacheEntry struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	ExpiresAt  time.Time
}

// ResponseCache is a simple in-memory TTL cache keyed by request path.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewResponseCache creates a ResponseCache with the given TTL.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get returns a cached entry for key, or nil if absent/expired.
func (c *ResponseCache) Get(key string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.ExpiresAt) {
		return nil
	}
	return e
}

// Set stores a response under key with the configured TTL.
func (c *ResponseCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry.ExpiresAt = time.Now().Add(c.ttl)
	c.entries[key] = entry
}

// Delete removes a single entry.
func (c *ResponseCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Flush removes all entries.
func (c *ResponseCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Len returns the number of entries (including expired ones not yet evicted).
func (c *ResponseCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// TTL returns the configured time-to-live for entries.
func (c *ResponseCache) TTL() time.Duration {
	return c.ttl
}
