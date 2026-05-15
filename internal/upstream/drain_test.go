package upstream

import (
	"testing"
)

func TestNewDrainRegistry_Empty(t *testing.T) {
	r := NewDrainRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Error("expected empty snapshot")
	}
}

func TestDrain_MarksBackend(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	if !r.IsDraining("http://backend:8080") {
		t.Error("expected backend to be draining")
	}
}

func TestRestore_RemovesBackend(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	r.Restore("http://backend:8080")
	if r.IsDraining("http://backend:8080") {
		t.Error("expected backend to no longer be draining")
	}
}

func TestAcquire_NotDraining_ReturnsNil(t *testing.T) {
	r := NewDrainRegistry()
	if err := r.Acquire("http://healthy:9000"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAcquire_Draining_ReturnsErr(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	err := r.Acquire("http://backend:8080")
	if err != ErrDrained {
		t.Errorf("expected ErrDrained, got %v", err)
	}
}

func TestAcquire_IncrementsActiveCount(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	_ = r.Acquire("http://backend:8080")
	_ = r.Acquire("http://backend:8080")
	if got := r.ActiveCount("http://backend:8080"); got != 2 {
		t.Errorf("expected active count 2, got %d", got)
	}
}

func TestRelease_DecrementsActiveCount(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	_ = r.Acquire("http://backend:8080")
	_ = r.Acquire("http://backend:8080")
	r.Release("http://backend:8080")
	if got := r.ActiveCount("http://backend:8080"); got != 1 {
		t.Errorf("expected active count 1, got %d", got)
	}
}

func TestActiveCount_Unknown_ReturnsZero(t *testing.T) {
	r := NewDrainRegistry()
	if got := r.ActiveCount("http://unknown:9999"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestSnapshot_ReflectsDrainedBackends(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://a:8001")
	r.Drain("http://b:8002")
	_ = r.Acquire("http://a:8001")

	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if snap["http://a:8001"] != 1 {
		t.Errorf("expected 1 active for a, got %d", snap["http://a:8001"])
	}
	if snap["http://b:8002"] != 0 {
		t.Errorf("expected 0 active for b, got %d", snap["http://b:8002"])
	}
}

func TestDrain_Idempotent(t *testing.T) {
	r := NewDrainRegistry()
	r.Drain("http://backend:8080")
	r.Drain("http://backend:8080") // should not panic or reset counter
	_ = r.Acquire("http://backend:8080")
	if got := r.ActiveCount("http://backend:8080"); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}
