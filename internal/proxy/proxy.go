package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/relayctl/internal/config"
)

// ReverseProxy wraps a set of backend proxies keyed by route prefix.
type ReverseProxy struct {
	mu      sync.RWMutex
	routes  map[string]*httputil.ReverseProxy
	cfg     *config.Config
}

// New creates a ReverseProxy from the given config.
func New(cfg *config.Config) (*ReverseProxy, error) {
	rp := &ReverseProxy{}
	if err := rp.reload(cfg); err != nil {
		return nil, err
	}
	return rp, nil
}

// Reload atomically replaces the routing table with a new config.
func (rp *ReverseProxy) Reload(cfg *config.Config) error {
	return rp.reload(cfg)
}

func (rp *ReverseProxy) reload(cfg *config.Config) error {
	routes := make(map[string]*httputil.ReverseProxy, len(cfg.Routes))
	for _, route := range cfg.Routes {
		target, err := url.Parse(route.Backend)
		if err != nil {
			return fmt.Errorf("invalid backend %q for route %q: %w", route.Backend, route.Prefix, err)
		}
		routes[route.Prefix] = httputil.NewSingleHostReverseProxy(target)
	}

	rp.mu.Lock()
	rp.routes = routes
	rp.cfg = cfg
	rp.mu.Unlock()

	log.Printf("proxy: loaded %d route(s)", len(routes))
	return nil
}

// ServeHTTP dispatches requests to the matching backend.
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	for prefix, backend := range rp.routes {
		if len(r.URL.Path) >= len(prefix) && r.URL.Path[:len(prefix)] == prefix {
			backend.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "no matching route", http.StatusBadGateway)
}
