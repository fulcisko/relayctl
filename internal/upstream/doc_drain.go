// Package upstream provides load balancing and traffic management primitives
// for relayctl's reverse proxy layer.
//
// # Drain Registry
//
// The DrainRegistry allows graceful removal of backends from rotation by
// marking them as "draining". A draining backend will not accept new
// connections but existing in-flight requests are tracked via a WaitGroup
// so callers can wait for them to complete before final shutdown.
//
// Usage:
//
//	reg := upstream.NewDrainRegistry()
//	reg.Drain("http://backend:8080")       // mark for draining
//	token, err := reg.Acquire("http://backend:8080") // returns err if draining
//	if err == nil {
//	    defer token.Release()
//	}
//	reg.Wait("http://backend:8080")        // block until all in-flight done
//	reg.Restore("http://backend:8080")     // re-enable backend
package upstream
