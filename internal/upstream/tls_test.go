package upstream

import (
	"testing"
)

func TestNewTLSRegistry_Empty(t *testing.T) {
	r := NewTLSRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	snap := r.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestTLSRegistry_Set_And_Get(t *testing.T) {
	r := NewTLSRegistry()
	cfg := TLSConfig{InsecureSkipVerify: true, ServerName: "example.com"}
	if err := r.Set("https://backend:443", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("https://backend:443")
	if !ok {
		t.Fatal("expected config to exist")
	}
	if got.ServerName != "example.com" {
		t.Errorf("expected ServerName=example.com, got %q", got.ServerName)
	}
	if !got.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify=true")
	}
}

func TestTLSRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewTLSRegistry()
	if err := r.Set("", TLSConfig{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestTLSRegistry_Transport_Present(t *testing.T) {
	r := NewTLSRegistry()
	_ = r.Set("https://backend:443", TLSConfig{ServerName: "svc"})
	tr := r.Transport("https://backend:443")
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestTLSRegistry_Transport_Missing(t *testing.T) {
	r := NewTLSRegistry()
	if tr := r.Transport("https://unknown"); tr != nil {
		t.Fatal("expected nil transport for unknown backend")
	}
}

func TestTLSRegistry_Delete(t *testing.T) {
	r := NewTLSRegistry()
	_ = r.Set("https://backend:443", TLSConfig{})
	r.Delete("https://backend:443")
	if _, ok := r.Get("https://backend:443"); ok {
		t.Fatal("expected config to be deleted")
	}
	if r.Transport("https://backend:443") != nil {
		t.Fatal("expected transport to be deleted")
	}
}

func TestTLSRegistry_Snapshot_IsCopy(t *testing.T) {
	r := NewTLSRegistry()
	_ = r.Set("https://a", TLSConfig{ServerName: "a"})
	_ = r.Set("https://b", TLSConfig{ServerName: "b"})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	// mutating snapshot must not affect registry
	delete(snap, "https://a")
	if _, ok := r.Get("https://a"); !ok {
		t.Fatal("registry should not be affected by snapshot mutation")
	}
}
