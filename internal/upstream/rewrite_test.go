package upstream

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewRewriteRegistry_Empty(t *testing.T) {
	r := NewRewriteRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestRewriteRegistry_Set_And_Get(t *testing.T) {
	r := NewRewriteRegistry()
	if err := r.Set("http://backend:8080", "/api", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rule, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if rule.Prefix != "/api" {
		t.Errorf("expected prefix /api, got %s", rule.Prefix)
	}
}

func TestRewriteRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewRewriteRegistry()
	if err := r.Set("", "/api", ""); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestRewriteRegistry_Delete(t *testing.T) {
	r := NewRewriteRegistry()
	_ = r.Set("http://backend:9000", "/v1", "/v2")
	r.Delete("http://backend:9000")
	if _, ok := r.Get("http://backend:9000"); ok {
		t.Fatal("expected rule to be deleted")
	}
}

func TestRewriteRegistry_Apply_RewritesPath(t *testing.T) {
	r := NewRewriteRegistry()
	_ = r.Set("http://svc:8080", "/api/v1", "/v1")
	req := &http.Request{URL: &url.URL{Path: "/api/v1/users"}}
	r.Apply("http://svc:8080", req)
	if req.URL.Path != "/v1/users" {
		t.Errorf("expected /v1/users, got %s", req.URL.Path)
	}
}

func TestRewriteRegistry_Apply_NoMatch(t *testing.T) {
	r := NewRewriteRegistry()
	_ = r.Set("http://svc:8080", "/api", "")
	req := &http.Request{URL: &url.URL{Path: "/other/path"}}
	r.Apply("http://svc:8080", req)
	if req.URL.Path != "/other/path" {
		t.Errorf("path should be unchanged, got %s", req.URL.Path)
	}
}

func TestRewriteRegistry_Apply_NoRule(t *testing.T) {
	r := NewRewriteRegistry()
	req := &http.Request{URL: &url.URL{Path: "/some/path"}}
	r.Apply("http://unknown:9999", req)
	if req.URL.Path != "/some/path" {
		t.Errorf("path should be unchanged, got %s", req.URL.Path)
	}
}

func TestRewriteRegistry_Snapshot_ReturnsCopy(t *testing.T) {
	r := NewRewriteRegistry()
	_ = r.Set("http://a:1", "/x", "/y")
	snap := r.Snapshot()
	delete(snap, "http://a:1")
	if _, ok := r.Get("http://a:1"); !ok {
		t.Fatal("original registry should not be affected by snapshot mutation")
	}
}
