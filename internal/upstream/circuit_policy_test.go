package upstream

import (
	"testing"
)

func TestNewCircuitPolicyRegistry_Empty(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestCircuitPolicyRegistry_Set_And_Get(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	p := CircuitPolicy{FailureThreshold: 3, SuccessThreshold: 2, TimeoutSeconds: 10}
	if err := r.Set("http://backend:8080", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected policy to be found")
	}
	if got != p {
		t.Errorf("expected %+v, got %+v", p, got)
	}
}

func TestCircuitPolicyRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	err := r.Set("", CircuitPolicy{FailureThreshold: 1, SuccessThreshold: 1, TimeoutSeconds: 5})
	if err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestCircuitPolicyRegistry_Set_InvalidThresholds(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	cases := []CircuitPolicy{
		{FailureThreshold: 0, SuccessThreshold: 1, TimeoutSeconds: 5},
		{FailureThreshold: 1, SuccessThreshold: 0, TimeoutSeconds: 5},
		{FailureThreshold: 1, SuccessThreshold: 1, TimeoutSeconds: 0},
		{FailureThreshold: 1, SuccessThreshold: 1, TimeoutSeconds: -1},
	}
	for _, p := range cases {
		if err := r.Set("http://b:8080", p); err == nil {
			t.Errorf("expected error for policy %+v", p)
		}
	}
}

func TestCircuitPolicyRegistry_Delete(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	p := CircuitPolicy{FailureThreshold: 2, SuccessThreshold: 1, TimeoutSeconds: 30}
	_ = r.Set("http://b:9090", p)
	r.Delete("http://b:9090")
	_, ok := r.Get("http://b:9090")
	if ok {
		t.Fatal("expected policy to be deleted")
	}
}

func TestCircuitPolicyRegistry_Snapshot(t *testing.T) {
	r := NewCircuitPolicyRegistry()
	_ = r.Set("http://a:1", CircuitPolicy{FailureThreshold: 5, SuccessThreshold: 2, TimeoutSeconds: 15})
	_ = r.Set("http://b:2", CircuitPolicy{FailureThreshold: 3, SuccessThreshold: 1, TimeoutSeconds: 20})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	// Mutating snapshot should not affect registry
	delete(snap, "http://a:1")
	if _, ok := r.Get("http://a:1"); !ok {
		t.Fatal("registry was mutated by snapshot modification")
	}
}
