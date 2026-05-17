package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHedgeRegistry_Empty(t *testing.T) {
	r := NewHedgeRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestHedgeRegistry_Set_And_Get(t *testing.T) {
	r := NewHedgeRegistry()
	cfg := HedgeConfig{Delay: 20 * time.Millisecond, MaxHedges: 2}
	if err := r.Set("http://backend:8080", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.MaxHedges != 2 {
		t.Errorf("expected MaxHedges=2, got %d", got.MaxHedges)
	}
}

func TestHedgeRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewHedgeRegistry()
	if err := r.Set("", HedgeConfig{Delay: 10 * time.Millisecond}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestHedgeRegistry_Set_DefaultsDelay(t *testing.T) {
	r := NewHedgeRegistry()
	_ = r.Set("http://b:9", HedgeConfig{})
	got, _ := r.Get("http://b:9")
	if got.Delay != 50*time.Millisecond {
		t.Errorf("expected default delay 50ms, got %v", got.Delay)
	}
	if got.MaxHedges != 1 {
		t.Errorf("expected default MaxHedges=1, got %d", got.MaxHedges)
	}
}

func TestHedgeRegistry_Delete(t *testing.T) {
	r := NewHedgeRegistry()
	_ = r.Set("http://x:1", HedgeConfig{Delay: 10 * time.Millisecond})
	r.Delete("http://x:1")
	if _, ok := r.Get("http://x:1"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestHedgeRegistry_Middleware_NoConfig_PassesThrough(t *testing.T) {
	r := NewHedgeRegistry()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := r.Middleware("http://unknown:80", next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestHedgeRegistry_Middleware_WithConfig(t *testing.T) {
	r := NewHedgeRegistry()
	_ = r.Set("http://b:80", HedgeConfig{Delay: 5 * time.Millisecond, MaxHedges: 1})
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := r.Middleware("http://b:80", next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("expected next handler to be called")
	}
}
