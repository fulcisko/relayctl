package circuitbreaker

import (
	"testing"
)

func defaultRegistryConfig() Config {
	return Config{MaxFailures: 3, OpenTimeout: 5}
}

func TestRegistry_GetCreatesBreaker(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	cb := r.Get("http://backend1")
	if cb == nil {
		t.Fatal("expected non-nil circuit breaker")
	}
}

func TestRegistry_GetReturnsSameInstance(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	cb1 := r.Get("http://backend1")
	cb2 := r.Get("http://backend1")
	if cb1 != cb2 {
		t.Fatal("expected same circuit breaker instance")
	}
}

func TestRegistry_GetDifferentKeys(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	cb1 := r.Get("http://backend1")
	cb2 := r.Get("http://backend2")
	if cb1 == cb2 {
		t.Fatal("expected different circuit breaker instances")
	}
}

func TestRegistry_Reset(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	cb := r.Get("http://backend1")
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected Open state, got %s", cb.State())
	}
	r.Reset("http://backend1")
	if cb.State() != StateClosed {
		t.Fatalf("expected Closed state after reset, got %s", cb.State())
	}
}

func TestRegistry_Reset_UnknownKey(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	// Should not panic on unknown key.
	r.Reset("http://unknown")
}

func TestRegistry_Snapshot(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	r.Get("http://a")
	r.Get("http://b")
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	for k, v := range snap {
		if v != StateClosed {
			t.Errorf("key %s: expected Closed, got %s", k, v)
		}
	}
}

func TestRegistry_Keys(t *testing.T) {
	r := NewRegistry(defaultRegistryConfig())
	r.Get("http://x")
	r.Get("http://y")
	keys := r.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}
