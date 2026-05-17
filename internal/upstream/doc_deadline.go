// Package upstream provides balancer and registry implementations
// for managing backend upstreams in relayctl.
//
// # Deadline Registry
//
// DeadlineRegistry allows per-backend request deadline configuration.
// A deadline sets an absolute time limit for the entire request lifecycle,
// distinct from per-phase timeouts in TimeoutRegistry.
//
// Example usage:
//
//	reg := upstream.NewDeadlineRegistry()
//	reg.Set("http://backend:8080", 5*time.Second)
//
//	deadline, ok := reg.Get("http://backend:8080")
//	if ok {
//	    ctx, cancel := context.WithTimeout(r.Context(), deadline)
//	    defer cancel()
//	    r = r.WithContext(ctx)
//	}
package upstream
