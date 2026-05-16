// Package upstream provides per-backend rate limit registry support.
//
// # Per-Backend Rate Limiting
//
// RateLimitRegistry allows operators to configure independent rate limits for
// each upstream backend. This is useful when backends have different capacity
// characteristics or SLAs.
//
// Example usage:
//
//	reg := upstream.NewRateLimitRegistry()
//	_ = reg.Set("http://api-server:8080", upstream.PerBackendRateLimit{
//		RequestsPerSecond: 100,
//		Burst:             200,
//	})
//
//	cfg, ok := reg.Get("http://api-server:8080")
//	if ok {
//		// apply cfg to a token-bucket limiter for this backend
//	}
//
// The registry is safe for concurrent use. Entries can be updated or removed
// at runtime without restarting the proxy.
package upstream
