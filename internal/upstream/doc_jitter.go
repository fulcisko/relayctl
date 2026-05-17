// Package upstream provides upstream balancer strategies and registries
// for relayctl's reverse proxy engine.
//
// # Jitter Registry
//
// JitterRegistry adds randomized delay to upstream requests on a per-backend
// basis. This is useful for spreading load spikes and avoiding thundering herd
// problems when many clients reconnect simultaneously.
//
// Each backend can be configured with a base delay and a maximum jitter window.
// The actual delay applied to each request is:
//
//	actual = base + rand(0, jitter)
//
// Example usage:
//
//	reg := upstream.NewJitterRegistry()
//	reg.Set("http://backend:8080", 10*time.Millisecond, 50*time.Millisecond)
//
//	delay := reg.Delay("http://backend:8080")
//	time.Sleep(delay)
package upstream
