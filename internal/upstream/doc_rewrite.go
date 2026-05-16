// Package upstream provides load-balancing and traffic-shaping primitives
// used by the relayctl reverse proxy.
//
// # Path Rewrite
//
// RewriteRegistry maps backend URLs to RewriteRule values. Each rule defines
// a Prefix that, when matched at the start of the request path, is replaced
// with Replacement before the request is forwarded.
//
// Example — strip /api/v1 prefix before forwarding to a backend:
//
//	reg := upstream.NewRewriteRegistry()
//	_ = reg.Set("http://svc:8080", "/api/v1", "")
//
//	// Inside your proxy handler:
//	reg.Apply("http://svc:8080", req)
//
// The registry is safe for concurrent reads but should be updated only
// via the admin API or at configuration reload time.
package upstream
