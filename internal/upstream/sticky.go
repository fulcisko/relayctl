// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"net/http"
	"sync"
)

// StickyBalancer routes requests to the same backend based on a session cookie.
// If no cookie is present or the backend is unavailable, it falls back to the
// underlying balancer.
type StickyBalancer struct {
	mu       sync.RWMutex
	balancer Balancer
	cookieName string
	sessions map[string]string // cookie value -> backend URL
}

// NewStickyBalancer wraps a Balancer with sticky session support using the
// given cookie name.
func NewStickyBalancer(b Balancer, cookieName string) *StickyBalancer {
	return &StickyBalancer{
		balancer:   b,
		cookieName: cookieName,
		sessions:   make(map[string]string),
	}
}

// Next returns the backend URL for the given request. If the request carries a
// known session cookie, the previously assigned backend is returned. Otherwise
// the underlying balancer picks a backend and the mapping is stored.
func (s *StickyBalancer) Next(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(s.cookieName)
	if err == nil && cookie.Value != "" {
		s.mu.RLock()
		backend, ok := s.sessions[cookie.Value]
		s.mu.RUnlock()
		if ok {
			return backend, true
		}
	}

	backend, ok := s.balancer.Next()
	if !ok {
		return "", false
	}

	if err == nil && cookie.Value != "" {
		s.mu.Lock()
		s.sessions[cookie.Value] = backend
		s.mu.Unlock()
	}

	return backend, true
}

// Forget removes a session mapping for the given cookie value.
func (s *StickyBalancer) Forget(cookieValue string) {
	s.mu.Lock()
	delete(s.sessions, cookieValue)
	s.mu.Unlock()
}

// SessionCount returns the number of active session mappings.
func (s *StickyBalancer) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
