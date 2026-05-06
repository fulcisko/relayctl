package middleware

import (
	"log"
	"net/http"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture the status code.
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader captures the status code before delegating to the underlying ResponseWriter.
func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger returns an HTTP middleware that logs each request with method,
// path, status code, and elapsed time.
//
// Example log output:
//
//	GET /api/relays 200 1.23ms
func Logger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &ResponseRecorder{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			logger.Printf("%s %s %d %s",
				r.Method,
				r.URL.Path,
				rec.StatusCode,
				time.Since(start),
			)
		})
	}
}
