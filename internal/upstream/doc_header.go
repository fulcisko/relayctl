// Package upstream provides balancer implementations and per-backend
// configuration registries used by the reverse proxy.
//
// # Header Registry
//
// NewHeaderRegistry creates a registry that stores request/response header
// overrides on a per-backend basis. Call Apply to mutate an *http.Request
// before it is forwarded.
//
// Example:
//
//	reg := upstream.NewHeaderRegistry()
//	reg.Set("http://backend:8080", upstream.HeaderRules{
//		RequestSet:  map[string]string{"X-Forwarded-By": "relayctl"},
//		RequestDel:  []string{"X-Internal-Token"},
//		ResponseSet: map[string]string{"X-Cache": "miss"},
//	})
//
//	// later, in the proxy transport:
//	reg.Apply("http://backend:8080", req)
package upstream
