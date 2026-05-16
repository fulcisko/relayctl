// Package upstream provides load balancing strategies for relayctl.
//
// # Priority Balancer
//
// The PriorityBalancer selects backends based on an ordered priority list.
// Backends with higher priority are always preferred. If a high-priority
// backend is unhealthy (as reported by a HealthChecker), the balancer
// falls back to the next available backend by priority.
//
// Usage:
//
//	groups := []upstream.PriorityGroup{
//	    {Backend: "http://primary:8080", Priority: 10},
//	    {Backend: "http://secondary:8080", Priority: 5},
//	    {Backend: "http://tertiary:8080", Priority: 1},
//	}
//	balancer := upstream.NewPriorityBalancer(groups, healthChecker)
//	next := balancer.Next(req)
//
// If all backends are unhealthy, Next returns an empty string.
package upstream
