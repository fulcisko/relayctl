package upstream

import (
	"net/http"
	"testing"
)

func TestNewTracingRegistry_Empty(t *testing.T) {
	reg := NewTracingRegistry()
	if len(reg.Snapshot()) != 0 {
		t.Fatal("expected empty registry")
	}
}

func TestTracingRegistry_Set_And_Get(t *testing.T) {
	reg := NewTracingRegistry()
	cfg := TracingConfig{InjectRequestID: true, HeaderName: "X-Trace"}
	if err := reg.Set("http://backend:9090", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := reg.Get("http://backend:9090")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.HeaderName != "X-Trace" {
		t.Errorf("expected X-Trace, got %s", got.HeaderName)
	}
}

func TestTracingRegistry_Set_EmptyBackend(t *testing.T) {
	reg := NewTracingRegistry()
	if err := reg.Set("", TracingConfig{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestTracingRegistry_Set_DefaultsHeaderName(t *testing.T) {
	reg := NewTracingRegistry()
	_ = reg.Set("http://backend:9090", TracingConfig{InjectRequestID: true})
	got, _ := reg.Get("http://backend:9090")
	if got.HeaderName != "X-Request-ID" {
		t.Errorf("expected default X-Request-ID, got %s", got.HeaderName)
	}
}

func TestTracingRegistry_Delete(t *testing.T) {
	reg := NewTracingRegistry()
	_ = reg.Set("http://backend:9090", TracingConfig{InjectRequestID: true})
	reg.Delete("http://backend:9090")
	if _, ok := reg.Get("http://backend:9090"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestTracingRegistry_Apply_InjectsHeader(t *testing.T) {
	reg := NewTracingRegistry()
	_ = reg.Set("http://backend:9090", TracingConfig{InjectRequestID: true, HeaderName: "X-Request-ID"})
	req, _ := http.NewRequest("GET", "/", nil)
	reg.Apply("http://backend:9090", req)
	if req.Header.Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set")
	}
}

func TestTracingRegistry_Apply_PreservesExisting(t *testing.T) {
	reg := NewTracingRegistry()
	_ = reg.Set("http://backend:9090", TracingConfig{InjectRequestID: true, HeaderName: "X-Request-ID"})
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	reg.Apply("http://backend:9090", req)
	if req.Header.Get("X-Request-ID") != "existing-id" {
		t.Error("expected existing X-Request-ID to be preserved")
	}
}

func TestTracingRegistry_Apply_NoConfig(t *testing.T) {
	reg := NewTracingRegistry()
	req, _ := http.NewRequest("GET", "/", nil)
	reg.Apply("http://unknown:9090", req)
	if req.Header.Get("X-Request-ID") != "" {
		t.Error("expected no header injection for unconfigured backend")
	}
}

func TestTracingRegistry_Snapshot(t *testing.T) {
	reg := NewTracingRegistry()
	_ = reg.Set("http://a:1", TracingConfig{InjectRequestID: true})
	_ = reg.Set("http://b:2", TracingConfig{InjectRequestID: false})
	snap := reg.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}
