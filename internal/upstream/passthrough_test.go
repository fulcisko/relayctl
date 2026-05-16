package upstream

import (
	"testing"
)

func TestNewPassthroughRegistry_Empty(t *testing.T) {
	reg := NewPassthroughRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if got := reg.Snapshot(); len(got) != 0 {
		t.Fatalf("expected empty snapshot, got %v", got)
	}
}

func TestPassthroughRegistry_Set_And_IsPassthrough(t *testing.T) {
	reg := NewPassthroughRegistry()
	if err := reg.Set("http://svc:8080"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reg.IsPassthrough("http://svc:8080") {
		t.Error("expected backend to be passthrough")
	}
}

func TestPassthroughRegistry_Set_EmptyBackend(t *testing.T) {
	reg := NewPassthroughRegistry()
	if err := reg.Set(""); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestPassthroughRegistry_IsPassthrough_Unknown(t *testing.T) {
	reg := NewPassthroughRegistry()
	if reg.IsPassthrough("http://unknown:9000") {
		t.Error("expected false for unregistered backend")
	}
}

func TestPassthroughRegistry_Delete(t *testing.T) {
	reg := NewPassthroughRegistry()
	_ = reg.Set("http://svc:8080")
	reg.Delete("http://svc:8080")
	if reg.IsPassthrough("http://svc:8080") {
		t.Error("expected backend to be removed")
	}
}

func TestPassthroughRegistry_Snapshot(t *testing.T) {
	reg := NewPassthroughRegistry()
	_ = reg.Set("http://a:1")
	_ = reg.Set("http://b:2")
	snap := reg.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}

func TestPassthroughRegistry_Snapshot_IsCopy(t *testing.T) {
	reg := NewPassthroughRegistry()
	_ = reg.Set("http://a:1")
	snap := reg.Snapshot()
	snap[0] = "mutated"
	if !reg.IsPassthrough("http://a:1") {
		t.Error("snapshot mutation should not affect registry")
	}
}
