package upstream

import (
	"net/http"
	"testing"
)

func TestNewHeaderRegistry_Empty(t *testing.T) {
	r := NewHeaderRegistry()
	snap := r.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestHeaderRegistry_Set_And_Get(t *testing.T) {
	r := NewHeaderRegistry()
	rule := HeaderRule{
		Set:    map[string]string{"X-Forwarded-By": "relayctl"},
		Remove: []string{"X-Internal"},
	}
	if err := r.Set("http://backend:8080", rule); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if got.Set["X-Forwarded-By"] != "relayctl" {
		t.Errorf("unexpected Set value: %v", got.Set)
	}
}

func TestHeaderRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewHeaderRegistry()
	if err := r.Set("", HeaderRule{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestHeaderRegistry_Delete(t *testing.T) {
	r := NewHeaderRegistry()
	_ = r.Set("http://backend:9000", HeaderRule{Set: map[string]string{"X-Env": "prod"}})
	r.Delete("http://backend:9000")
	_, ok := r.Get("http://backend:9000")
	if ok {
		t.Fatal("expected rule to be deleted")
	}
}

func TestHeaderRegistry_Apply_SetsHeaders(t *testing.T) {
	r := NewHeaderRegistry()
	_ = r.Set("http://svc:80", HeaderRule{
		Set:    map[string]string{"X-Version": "2"},
		Add:    map[string]string{"X-Tag": "canary"},
		Remove: []string{"X-Legacy"},
	})
	req, _ := http.NewRequest(http.MethodGet, "http://svc:80/path", nil)
	req.Header.Set("X-Legacy", "old")
	r.Apply("http://svc:80", req)
	if req.Header.Get("X-Version") != "2" {
		t.Errorf("expected X-Version=2, got %q", req.Header.Get("X-Version"))
	}
	if req.Header.Get("X-Tag") != "canary" {
		t.Errorf("expected X-Tag=canary, got %q", req.Header.Get("X-Tag"))
	}
	if req.Header.Get("X-Legacy") != "" {
		t.Errorf("expected X-Legacy removed, got %q", req.Header.Get("X-Legacy"))
	}
}

func TestHeaderRegistry_Apply_NoRule(t *testing.T) {
	r := NewHeaderRegistry()
	req, _ := http.NewRequest(http.MethodGet, "http://unknown:80/", nil)
	req.Header.Set("X-Keep", "yes")
	r.Apply("http://unknown:80", req)
	if req.Header.Get("X-Keep") != "yes" {
		t.Error("header should be unchanged when no rule exists")
	}
}

func TestHeaderRegistry_Snapshot_ReturnsCopy(t *testing.T) {
	r := NewHeaderRegistry()
	_ = r.Set("http://a:1", HeaderRule{Set: map[string]string{"X-A": "1"}})
	snap := r.Snapshot()
	delete(snap, "http://a:1")
	_, ok := r.Get("http://a:1")
	if !ok {
		t.Error("snapshot mutation should not affect registry")
	}
}
