package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relayctl/internal/config"
)

func TestNewServer(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Addr: ":0",
		Routes: []config.Route{
			{Prefix: "/", Backend: backend.URL},
		},
	}

	rp, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	srv := NewServer(cfg.Addr, rp)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServer_Shutdown(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Addr: "127.0.0.1:0",
		Routes: []config.Route{
			{Prefix: "/", Backend: backend.URL},
		},
	}

	rp, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	srv := NewServer(cfg.Addr, rp)

	go func() {
		_ = srv.ListenAndServe()
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}

func TestServer_UpdateAddr(t *testing.T) {
	rp, _ := New(&config.Config{
		Addr:   ":0",
		Routes: []config.Route{},
	})
	srv := NewServer(":8080", rp)
	srv.UpdateAddr(":9090")
	if srv.httpServer.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", srv.httpServer.Addr)
	}
}
