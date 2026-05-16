// Package upstream provides upstream backend management for relayctl.
//
// # Request Tracing
//
// The TracingRegistry allows per-backend tracing configuration, enabling
// injection of trace headers (e.g. X-Request-ID, X-Trace-ID) into proxied
// requests before they reach the upstream.
//
// Usage:
//
//	reg := upstream.NewTracingRegistry()
//	reg.Set("http://backend:8080", upstream.TracingConfig{
//		InjectRequestID: true,
//		HeaderName:      "X-Request-ID",
//	})
//
//	headers := reg.Headers("http://backend:8080", existingRequestID)
//
package upstream
