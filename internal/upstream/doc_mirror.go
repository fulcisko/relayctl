// Package upstream provides a suite of load-balancing strategies for
// distributing traffic across backend servers.
//
// # Mirror Balancer
//
// MirrorBalancer wraps two Balancer instances — a primary and a shadow — and
// forwards all live traffic to the primary while asynchronously replicating
// each request to the shadow backend.
//
// Shadow responses are silently discarded; the balancer is designed for
// dark-launch testing, traffic recording, and canary analysis without
// impacting production latency.
//
// Example usage:
//
//	primary := upstream.New([]string{"prod-a:8080", "prod-b:8080"})
//	shadow  := upstream.New([]string{"staging:8080"})
//
//	mirror, err := upstream.NewMirrorBalancer(primary, shadow)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// In your reverse-proxy handler:
//	//   backend, _ := mirror.Next(r)
//	//   mirror.Mirror(r)   // fire-and-forget shadow copy
//
// Mirroring can be toggled at runtime via the admin API:
//
//	PUT /admin/mirror  {"enabled": false}
//	GET /admin/mirror
package upstream
