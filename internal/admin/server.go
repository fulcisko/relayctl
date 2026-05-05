package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server is a lightweight HTTP server for the admin interface.
type Server struct {
	addr    string
	handler *Handler
	httpSrv *http.Server
}

// NewServer creates an admin Server bound to addr.
func NewServer(addr string, handler *Handler) *Server {
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return &Server{
		addr:    addr,
		handler: handler,
		httpSrv: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Start begins listening in a goroutine and returns any immediate bind error.
func (s *Server) Start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("admin server: %w", err)
		}
	case <-time.After(50 * time.Millisecond):
		// server started successfully
	}
	return nil
}

// Shutdown gracefully stops the admin server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
