package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter tracks request counts per IP using a sliding window.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	max      int
	window   time.Duration
}

type bucket struct {
	count     int
	resetAt   time.Time
}

// New creates a Limiter that allows max requests per window per IP.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		max:     max,
		window:  window,
	}
}

// Allow returns true if the given key (e.g. IP) is within the rate limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}

	if b.count >= l.max {
		return false
	}
	b.count++
	return true
}

// Middleware returns an HTTP middleware that enforces rate limiting by remote IP.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ipFromRequest(r)
		if !l.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ipFromRequest(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
