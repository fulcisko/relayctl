// Package upstream provides circuit policy registry for per-backend circuit breaker configuration.
//
// CircuitPolicyRegistry allows operators to define per-backend circuit breaker
// thresholds independently of the global circuit breaker settings.
//
// Example usage:
//
//	reg := upstream.NewCircuitPolicyRegistry()
//	err := reg.Set("http://backend:8080", upstream.CircuitPolicy{
//		Threshold: 5,
//		Timeout:   30 * time.Second,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	policy, ok := reg.Get("http://backend:8080")
package upstream
