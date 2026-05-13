// Package upstream provides a round-robin load balancer for managing
// a pool of backend URLs. It supports atomic updates to the backend
// list and is safe for concurrent use.
//
// Example:
//
//	b, err := upstream.New([]string{
//		"http://backend1:8080",
//		"http://backend2:8080",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	next, _ := b.Next()
//	fmt.Println(next) // http://backend1:8080
package upstream
