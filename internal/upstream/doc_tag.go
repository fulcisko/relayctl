// Package upstream provides balancer and registry primitives used by the
// reverse proxy to route, shape, and observe traffic to backend services.
//
// # Tag Registry
//
// TagRegistry associates an arbitrary set of string tags with each backend
// URL. Tags are free-form labels (e.g. "canary", "eu-west", "v2") that
// higher-level routing logic can inspect to make forwarding decisions.
//
// Example usage:
//
//	reg := upstream.NewTagRegistry()
//	reg.Set("http://backend-a:8080", []string{"stable", "eu"})
//
//	tags, ok := reg.Get("http://backend-a:8080")
//	if ok {
//	    fmt.Println(tags) // [stable eu]
//	}
//
// Tags are stored and returned as a defensive copy so callers cannot mutate
// the internal state.
package upstream
