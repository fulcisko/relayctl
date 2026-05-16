// Package upstream provides load balancing and upstream management
// for the relayctl reverse proxy.
//
// # Header Registry
//
// The HeaderRegistry allows per-backend request and response header
// manipulation. Headers can be injected, overwritten, or removed before
// a request is forwarded or after a response is received.
//
// Example usage:
//
//	reg := upstream.NewHeaderRegistry()
//	reg.Set("http://backend:8080", upstream.HeaderRules{
//		RequestAdd:  map[string]string{"X-Forwarded-By": "relayctl"},
//		ResponseDel: []string{"X-Internal-Token"},
//	})
//
//	// Apply to an outgoing request:
//	reg.Apply("http://backend:8080", req)
package upstream
