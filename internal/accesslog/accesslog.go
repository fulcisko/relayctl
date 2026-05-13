package accesslog

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Entry represents a single access log record.
type Entry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	ClientIP   string `json:"client_ip"`
	Backend    string `json:"backend"`
}

// Logger writes structured JSON access log entries.
type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

// New creates a new access Logger writing to w.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{out: w}
}

// Log writes an Entry to the underlying writer.
func (l *Logger) Log(e Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return json.NewEncoder(l.out).Encode(e)
}

// Middleware returns an http.Handler that logs each request.
func (l *Logger) Middleware(backend string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		_ = l.Log(Entry{
			Timestamp:  start.UTC().Format(time.RFC3339),
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     rw.status,
			DurationMs: time.Since(start).Milliseconds(),
			ClientIP:   r.RemoteAddr,
			Backend:    backend,
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
