// Package upstream — GeoBalancer
//
// GeoBalancer routes each request to a region-specific child balancer based on
// the client's IP address. If no region can be determined, or no balancer is
// registered for that region, the request falls through to a configurable
// fallback balancer.
//
// Usage:
//
//	// Build per-region balancers.
//	usB, _ := upstream.New([]string{"http://us-east:8080", "http://us-west:8080"})
//	euB, _ := upstream.New([]string{"http://eu-central:8080"})
//	fb,  _ := upstream.New([]string{"http://global:8080"})
//
//	// Provide a resolver that maps IPs to region strings.
//	resolve := func(ip string) string {
//	    if strings.HasPrefix(ip, "10.") { return "us" }
//	    return "eu"
//	}
//
//	geo, err := upstream.NewGeoBalancer(
//	    map[string]upstream.Balancer{"us": usB, "eu": euB},
//	    fb,
//	    resolve,
//	)
//
//	// At runtime, add or replace a region's balancer.
//	apB, _ := upstream.New([]string{"http://ap-south:8080"})
//	geo.SetRegion("ap", apB)
//
//	// Remove a region's balancer, falling back to the default.
//	geo.RemoveRegion("ap")
//
//	// List all currently registered region keys.
//	regions := geo.Regions() // e.g. ["us", "eu"]
//
// The admin endpoint /admin/geo exposes a GET (snapshot) and PUT (update
// region backends) interface for live configuration.
package upstream
