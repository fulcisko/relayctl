// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"errors"
	"net"
	"net/http"
	"sync"
)

// RegionResolver maps an IP address to a region string.
type RegionResolver func(ip string) string

// GeoBalancer routes requests to region-specific balancers, falling back to a
// default balancer when no region match is found.
type GeoBalancer struct {
	mu       sync.RWMutex
	regions  map[string]Balancer
	fallback Balancer
	resolve  RegionResolver
}

// NewGeoBalancer creates a GeoBalancer. fallback must not be nil.
func NewGeoBalancer(regions map[string]Balancer, fallback Balancer, resolve RegionResolver) (*GeoBalancer, error) {
	if fallback == nil {
		return nil, errors.New("geo balancer: fallback balancer must not be nil")
	}
	if resolve == nil {
		return nil, errors.New("geo balancer: region resolver must not be nil")
	}
	r := make(map[string]Balancer, len(regions))
	for k, v := range regions {
		if v == nil {
			return nil, errors.New("geo balancer: nil balancer for region " + k)
		}
		r[k] = v
	}
	return &GeoBalancer{regions: r, fallback: fallback, resolve: resolve}, nil
}

// Next picks a backend based on the X-Forwarded-For or RemoteAddr of the request.
func (g *GeoBalancer) Next(r *http.Request) (string, error) {
	ip := extractIP(r)
	region := g.resolve(ip)

	g.mu.RLock()
	b, ok := g.regions[region]
	g.mu.RUnlock()

	if ok {
		return b.Next(r)
	}
	return g.fallback.Next(r)
}

// Backends returns the union of all backends across regions and the fallback.
func (g *GeoBalancer) Backends() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	seen := map[string]struct{}{}
	var out []string
	for _, b := range g.regions {
		for _, addr := range b.Backends() {
			if _, exists := seen[addr]; !exists {
				seen[addr] = struct{}{}
				out = append(out, addr)
			}
		}
	}
	for _, addr := range g.fallback.Backends() {
		if _, exists := seen[addr]; !exists {
			out = append(out, addr)
		}
	}
	return out
}

// SetRegion replaces (or adds) the balancer for the given region.
func (g *GeoBalancer) SetRegion(region string, b Balancer) error {
	if b == nil {
		return errors.New("geo balancer: balancer must not be nil")
	}
	g.mu.Lock()
	g.regions[region] = b
	g.mu.Unlock()
	return nil
}

// extractIP returns the client IP from X-Forwarded-For or RemoteAddr.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
