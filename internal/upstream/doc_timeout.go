// Package upstream provides load balancing and upstream management
// for the relayctl reverse proxy.
//
// # Per-Backend Timeout Registry
//
// The TimeoutRegistry allows configuring per-backend dial, response header,
// and idle connection timeouts independently of the global proxy settings.
//
// Example usage:
//
//	reg := upstream.NewTimeoutRegistry()
//	reg.Set("http://backend-a:8080", upstream.TimeoutConfig{
//		Dial:           2 * time.Second,
//		ResponseHeader: 10 * time.Second,
//		Idle:           30 * time.Second,
//	})
//
//	transport := reg.Transport("http://backend-a:8080")
//	// transport is a *http.Transport with the specified timeouts applied.
//
// If no entry exists for a backend, Transport returns nil and the caller
// should fall back to a default transport.
package upstream
