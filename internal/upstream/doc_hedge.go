// Package upstream — hedge registry
//
// HedgeRegistry provides per-backend hedged-request configuration.
//
// A hedged request is a speculative duplicate: if the primary upstream call
// has not returned within a configured delay, a second request is fired in
// parallel and whichever response arrives first is used. This reduces
// tail-latency at the cost of slightly increased backend load.
//
// Usage:
//
//	reg := upstream.NewHedgeRegistry()
//	_ = reg.Set("http://api:8080", upstream.HedgeConfig{
//		Delay:     30 * time.Millisecond,
//		MaxHedges: 1,
//	})
//
//	// Wrap a handler:
//	h := reg.Middleware("http://api:8080", myHandler)
//
// The admin API exposes this registry at /admin/hedge.
package upstream
