// Package upstream provides load balancing strategies for relayctl.
//
// # Tag Filter Balancer
//
// NewTagFilterBalancer wraps any Balancer and filters candidates by tag.
// Only backends whose tags (from a TagRegistry) intersect with the required
// set are eligible. If no backend matches the required tags the call falls
// back to the underlying balancer without filtering.
//
// Example usage:
//
//	tags := upstream.NewTagRegistry()
//	_ = tags.Set("http://backend1:8080", []string{"eu", "premium"})
//
//	base := upstream.New([]string{"http://backend1:8080", "http://backend2:8080"})
//	filtered := upstream.NewTagFilterBalancer(base, tags, []string{"eu"})
//
//	// Next will only return backend1 because it carries the "eu" tag.
//	backend, _ := filtered.Next(req)
package upstream
