// Package upstream provides a tag-based filter balancer that selects backends
// matching a required tag from a TagRegistry before delegating to an inner
// Balancer. If no backends carry the required tag the inner balancer is used
// as a fallback so traffic is never dropped.
//
// Usage:
//
//	tags := upstream.NewTagRegistry()
//	_ = tags.Set("http://backend1:8080", []string{"us-east", "premium"})
//	_ = tags.Set("http://backend2:8080", []string{"eu-west"})
//
//	inner, _ := upstream.New([]string{
//		"http://backend1:8080",
//		"http://backend2:8080",
//	})
//
//	filtered := upstream.NewTagFilterBalancer(inner, tags, "us-east")
//	backend, done := filtered.Next(req)
package upstream
