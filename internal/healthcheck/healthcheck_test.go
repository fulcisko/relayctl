package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegister_And_IsHealthy_Default(t *testing.T) {
	c := New(time.Second, time.Second)
	c.Register("http://example.com")
	if !c.IsHealthy("http://example.com") {
		t.Fatal("newly registered backend should be healthy by default")
	}
}

func TestIsHealthy_UnknownURL(t *testing.T) {
	c := New(time.Second, time.Second)
	if !c.IsHealthy("http://unknown") {
		t.Fatal("unknown backend should be assumed healthy")
	}
}

func TestProbe_HealthyBackend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New(100*time.Millisecond, time.Second)
	c.Register(ts.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	time.Sleep(250 * time.Millisecond)

	if !c.IsHealthy(ts.URL) {
		t.Fatal("backend returning 200 should be healthy")
	}
}

func TestProbe_UnhealthyBackend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := New(100*time.Millisecond, time.Second)
	c.Register(ts.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	time.Sleep(250 * time.Millisecond)

	if c.IsHealthy(ts.URL) {
		t.Fatal("backend returning 500 should be unhealthy")
	}
}

func TestProbe_UnreachableBackend(t *testing.T) {
	c := New(100*time.Millisecond, 100*time.Millisecond)
	c.Register("http://127.0.0.1:19999")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	time.Sleep(350 * time.Millisecond)

	if c.IsHealthy("http://127.0.0.1:19999") {
		t.Fatal("unreachable backend should be unhealthy")
	}
}

func TestSnapshot(t *testing.T) {
	c := New(time.Second, time.Second)
	c.Register("http://a.example.com")
	c.Register("http://b.example.com")

	snap := c.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(snap))
	}
}

func TestStop(t *testing.T) {
	c := New(50*time.Millisecond, 50*time.Millisecond)
	ctx := context.Background()
	c.Start(ctx)
	// Should not panic or block
	c.Stop()
}
