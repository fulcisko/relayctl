// Package upstream provides quota management for per-backend request quotas.
//
// # Quota Registry
//
// QuotaRegistry tracks per-backend request quotas with configurable windows.
// Each backend can be assigned a maximum number of requests allowed within
// a rolling time window.
//
// # Usage
//
//	reg := upstream.NewQuotaRegistry()
//	err := reg.Set("http://backend:8080", upstream.QuotaConfig{
//		MaxRequests: 1000,
//		Window:      time.Minute,
//	})
//
//	allowed, err := reg.Allow("http://backend:8080")
//	if !allowed {
//		// quota exceeded
//	}
package upstream
