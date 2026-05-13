// Package upstream provides load balancing and health-aware backend selection
// for the relayctl reverse proxy.
//
// FilteredBalancer wraps a round-robin Balancer with a HealthChecker to ensure
// that only healthy backends receive traffic. When all backends are unhealthy,
// Next returns an empty string and the proxy layer should respond with 503.
//
// Usage:
//
//	bal, _ := upstream.New([]string{"http://backend1:8080", "http://backend2:8080"})
//	hc := healthcheck.New(healthcheck.Config{...})
//	fb := upstream.NewFilteredBalancer(bal, hc)
//
//	// In the proxy handler:
//	target := fb.Next()
//	if target == "" {
//		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
//		return
//	}
package upstream
