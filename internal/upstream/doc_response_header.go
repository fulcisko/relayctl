// Package upstream provides pluggable load-balancing strategies and
// per-backend request/response manipulation primitives for relayctl.
//
// # Response Header Registry
//
// ResponseHeaderRegistry allows per-backend response header mutation.
// Headers can be added, overwritten, or deleted before the response
// is forwarded back to the client.
//
// Example usage:
//
//	reg := upstream.NewResponseHeaderRegistry()
//	reg.Set("http://backend:8080", upstream.ResponseHeaderRules{
//		Set:    map[string]string{"X-Powered-By": "relayctl"},
//		Delete: []string{"Server"},
//	})
//
//	// In your RoundTripper:
//	resp, err := next.RoundTrip(req)
//	if err == nil {
//		reg.Apply(backendURL, resp.Header)
//	}
package upstream
