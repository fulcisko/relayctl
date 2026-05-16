// Package upstream provides pluggable load-balancing strategies and
// upstream management primitives for relayctl.
//
// # Passthrough Registry
//
// PassthroughRegistry allows certain backends to be marked as "passthrough"
// targets — requests are forwarded without modification, bypassing middleware
// such as header injection, rewriting, or rate limiting.
//
// Use cases:
//   - Trusted internal services that must receive raw requests.
//   - Debug/tracing endpoints where header mutation would interfere.
//
// Example:
//
//	reg := upstream.NewPassthroughRegistry()
//	reg.Set("http://internal.svc:9000")
//	if reg.IsPassthrough("http://internal.svc:9000") {
//	    // skip middleware chain
//	}
package upstream
