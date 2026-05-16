// Package upstream provides load balancing and upstream management
// for the relayctl reverse proxy.
//
// # Response Cache
//
// ResponseCache provides a simple TTL-based in-memory cache for upstream
// HTTP responses. It is keyed by an arbitrary string (typically the
// request path + query) and stores the raw response body along with
// status code and headers.
//
// Usage:
//
//	cache := upstream.NewResponseCache(upstream.CacheOptions{
//		TTL:      30 * time.Second,
//		MaxItems: 512,
//	})
//
//	// Store a response
//	cache.Set("GET:/api/v1/status", entry)
//
//	// Retrieve a cached response
//	if entry, ok := cache.Get("GET:/api/v1/status"); ok {
//		// serve from cache
//	}
//
// The CacheMiddleware wraps an http.RoundTripper and transparently
// caches GET responses that return a 200 status code.
package upstream
