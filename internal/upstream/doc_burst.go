// Package upstream provides load balancing strategies for relayctl.
//
// # Burst Balancer
//
// NewBurstBalancer wraps a primary balancer and a burst balancer. When the
// number of in-flight requests exceeds a configurable threshold, overflow
// traffic is redirected to the burst pool instead of the primary pool.
//
// Example usage:
//
//	primary, _ := upstream.New([]string{"http://primary:8080"})
//	burst, _ := upstream.New([]string{"http://burst1:8080", "http://burst2:8080"})
//	bb, err := upstream.NewBurstBalancer(primary, burst, upstream.BurstOptions{
//		Threshold:   50,
//		MaxBurst:    200,
//		Cooldown:    5 * time.Second,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// When active connections exceed Threshold the balancer begins routing new
// requests to the burst pool. Once active connections drop below Threshold
// for at least Cooldown duration, routing returns to the primary pool.
package upstream
