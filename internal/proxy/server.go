package proxy

import (
	"context"
	"log"
	"net/http"
	"time"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
)

// Server wraps http.Server with a ReverseProxy handler.
type Server struct {
	httpServer *http.Server
	proxy      *ReverseProxy
}

// NewServer creates a Server listening on addr using the given ReverseProxy.
func NewServer(addr string, rp *ReverseProxy) *Server {
	s := &Server{proxy: rp}
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      rp,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return s
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	log.Printf("server: listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server with the provided context.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("server: shutting down")
	return s.httpServer.Shutdown(ctx)
}

// UpdateAddr replaces the listen address. The change takes effect on the next
// call to ListenAndServe (after a Shutdown/restart cycle).
func (s *Server) UpdateAddr(addr string) {
	s.httpServer.Addr = addr
}
