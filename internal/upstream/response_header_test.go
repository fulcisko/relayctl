package upstream

import (
	"net/http"
	"testing"
)

func TestNewResponseHeaderRegistry_Empty(t *testing.T) {
	r := NewResponseHeaderRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestResponseHeaderRegistry_Set_And_Get(t *testing.T) {
	r := NewResponseHeaderRegistry()
	rules := ResponseHeaderRules{
		Set: map[string]string{"X-Cache": "miss"},
		Del: []string{"Server"},
	}
	if err := r.Set("http://backend:9000", rules); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := r.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected rules to be present")
	}
	if got.Set["X-Cache"] != "miss" {
		t.Errorf("Set[X-Cache] = %q, want \"miss\"", got.Set["X-Cache"])
	}
	if len(got.Del) != 1 || got.Del[0] != "Server" {
		t.Errorf("Del = %v, want [Server]", got.Del)
	}
}

func TestResponseHeaderRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewResponseHeaderRegistry()
	if err := r.Set("", ResponseHeaderRules{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestResponseHeaderRegistry_Delete(t *testing.T) {
	r := NewResponseHeaderRegistry()
	_ = r.Set("http://backend:9000", ResponseHeaderRules{Set: map[string]string{"X-Foo": "bar"}})
	r.Delete("http://backend:9000")
	if _, ok := r.Get("http://backend:9000"); ok {
		t.Fatal("expected rules to be removed")
	}
}

func TestResponseHeaderRegistry_Apply_SetsAndDeletes(t *testing.T) {
	r := NewResponseHeaderRegistry()
	_ = r.Set("http://backend:9000", ResponseHeaderRules{
		Set: map[string]string{"X-Powered-By": "relayctl"},
		Del: []string{"Server"},
	})

	resp := &http.Response{
		Header: http.Header{
			"Server": []string{"nginx"},
		},
	}
	r.Apply("http://backend:9000", resp)

	if v := resp.Header.Get("X-Powered-By"); v != "relayctl" {
		t.Errorf("X-Powered-By = %q, want \"relayctl\"", v)
	}
	if v := resp.Header.Get("Server"); v != "" {
		t.Errorf("Server = %q, want deleted", v)
	}
}

func TestResponseHeaderRegistry_Apply_NoRules_Noop(t *testing.T) {
	r := NewResponseHeaderRegistry()
	resp := &http.Response{Header: http.Header{"X-Custom": []string{"keep"}}}
	r.Apply("http://unknown:9000", resp) // should not panic or modify
	if v := resp.Header.Get("X-Custom"); v != "keep" {
		t.Errorf("X-Custom = %q, want \"keep\"", v)
	}
}

func TestResponseHeaderRegistry_Apply_NilResponse(t *testing.T) {
	r := NewResponseHeaderRegistry()
	_ = r.Set("http://backend:9000", ResponseHeaderRules{Set: map[string]string{"X-Foo": "bar"}})
	// must not panic
	r.Apply("http://backend:9000", nil)
}

func TestResponseHeaderRegistry_Snapshot(t *testing.T) {
	r := NewResponseHeaderRegistry()
	_ = r.Set("http://a:1", ResponseHeaderRules{Set: map[string]string{"A": "1"}})
	_ = r.Set("http://b:2", ResponseHeaderRules{Set: map[string]string{"B": "2"}})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Errorf("snapshot len = %d, want 2", len(snap))
	}
}
