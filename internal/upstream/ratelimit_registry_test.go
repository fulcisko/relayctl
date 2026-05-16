package upstream

import (
	"testing"
)

func TestNewRateLimitRegistry_Empty(t *testing.T) {
	r := NewRateLimitRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Error("expected empty snapshot")
	}
}

func TestRateLimitRegistry_Set_And_Get(t *testing.T) {
	r := NewRateLimitRegistry()
	cfg := PerBackendRateLimit{RequestsPerSecond: 10, Burst: 20}
	if err := r.Set("http://backend:8080", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.RequestsPerSecond != 10 || got.Burst != 20 {
		t.Errorf("unexpected config: %+v", got)
	}
}

func TestRateLimitRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewRateLimitRegistry()
	if err := r.Set("", PerBackendRateLimit{RequestsPerSecond: 5, Burst: 1}); err == nil {
		t.Error("expected error for empty backend")
	}
}

func TestRateLimitRegistry_Set_InvalidRPS(t *testing.T) {
	r := NewRateLimitRegistry()
	if err := r.Set("http://backend:8080", PerBackendRateLimit{RequestsPerSecond: 0, Burst: 5}); err == nil {
		t.Error("expected error for zero rps")
	}
}

func TestRateLimitRegistry_Set_DefaultsBurst(t *testing.T) {
	r := NewRateLimitRegistry()
	if err := r.Set("http://backend:8080", PerBackendRateLimit{RequestsPerSecond: 5, Burst: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := r.Get("http://backend:8080")
	if got.Burst != 1 {
		t.Errorf("expected burst=1, got %d", got.Burst)
	}
}

func TestRateLimitRegistry_Delete(t *testing.T) {
	r := NewRateLimitRegistry()
	_ = r.Set("http://backend:8080", PerBackendRateLimit{RequestsPerSecond: 5, Burst: 10})
	r.Delete("http://backend:8080")
	if _, ok := r.Get("http://backend:8080"); ok {
		t.Error("expected entry to be deleted")
	}
}

func TestRateLimitRegistry_Snapshot(t *testing.T) {
	r := NewRateLimitRegistry()
	_ = r.Set("http://a:8080", PerBackendRateLimit{RequestsPerSecond: 1, Burst: 2})
	_ = r.Set("http://b:8080", PerBackendRateLimit{RequestsPerSecond: 3, Burst: 4})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
	snap["http://a:8080"] = PerBackendRateLimit{RequestsPerSecond: 99}
	got, _ := r.Get("http://a:8080")
	if got.RequestsPerSecond == 99 {
		t.Error("snapshot mutation affected registry")
	}
}
