// Package upstream provides upstream balancer strategies and registries
// for relayctl's reverse proxy engine.
//
// # Concurrency Limiter Registry
//
// The ConcurrencyRegistry tracks per-backend maximum in-flight request limits.
// When a backend exceeds its configured limit, Acquire returns an error and the
// proxy can shed load or fall back to another backend.
//
// Usage:
//
//	reg := upstream.NewConcurrencyRegistry()
//	_ = reg.Set("http://backend:8080", 100) // max 100 concurrent requests
//
//	tok, err := reg.Acquire("http://backend:8080")
//	if err != nil {
//	    // shed load
//	}
//	defer tok.Release()
package upstream
