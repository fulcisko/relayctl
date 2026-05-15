package upstream

import (
	"net/http"
	"testing"
	"time"
)

func TestNewTimeoutRegistry_Empty(t *testing.T) {
	r := NewTimeoutRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	snap := r.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestTimeoutRegistry_Set_And_Get(t *testing.T) {
	r := NewTimeoutRegistry()
	cfg := TimeoutConfig{
		Dial:           100 * time.Millisecond,
		ResponseHeader: 200 * time.Millisecond,
		Idle:           300 * time.Millisecond,
	}
	if err := r.Set("http://backend:8080", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := r.Get("http://backend:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cfg {
		t.Fatalf("expected %+v, got %+v", cfg, got)
	}
}

func TestTimeoutRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewTimeoutRegistry()
	if err := r.Set("", TimeoutConfig{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestTimeoutRegistry_Get_Missing(t *testing.T) {
	r := NewTimeoutRegistry()
	_, err := r.Get("http://missing:9090")
	if err != ErrNoTimeoutConfig {
		t.Fatalf("expected ErrNoTimeoutConfig, got %v", err)
	}
}

func TestTimeoutRegistry_Delete(t *testing.T) {
	r := NewTimeoutRegistry()
	_ = r.Set("http://backend:8080", TimeoutConfig{Dial: 50 * time.Millisecond})
	r.Delete("http://backend:8080")
	_, err := r.Get("http://backend:8080")
	if err != ErrNoTimeoutConfig {
		t.Fatal("expected entry to be deleted")
	}
}

func TestTimeoutRegistry_Transport_Present(t *testing.T) {
	r := NewTimeoutRegistry()
	_ = r.Set("http://backend:8080", TimeoutConfig{
		Dial:           10 * time.Millisecond,
		ResponseHeader: 20 * time.Millisecond,
		Idle:           30 * time.Millisecond,
	})
	tr := r.Transport("http://backend:8080")
	if tr == http.DefaultTransport {
		t.Fatal("expected custom transport, got DefaultTransport")
	}
}

func TestTimeoutRegistry_Transport_Missing(t *testing.T) {
	r := NewTimeoutRegistry()
	tr := r.Transport("http://unknown:9999")
	if tr != http.DefaultTransport {
		t.Fatal("expected DefaultTransport for missing backend")
	}
}

func TestTimeoutRegistry_Snapshot_Multiple(t *testing.T) {
	r := NewTimeoutRegistry()
	_ = r.Set("http://a:1", TimeoutConfig{Dial: 1 * time.Millisecond})
	_ = r.Set("http://b:2", TimeoutConfig{Dial: 2 * time.Millisecond})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if snap["http://a:1"].Dial != 1*time.Millisecond {
		t.Fatal("unexpected value for http://a:1")
	}
}
