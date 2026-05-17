package upstream

import (
	"testing"
	"time"
)

func TestNewDeadlineRegistry_Empty(t *testing.T) {
	reg := NewDeadlineRegistry()
	snap := reg.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestDeadlineRegistry_Set_And_Get(t *testing.T) {
	reg := NewDeadlineRegistry()
	if err := reg.Set("http://backend:9000", 3*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if d != 3*time.Second {
		t.Fatalf("expected 3s, got %v", d)
	}
}

func TestDeadlineRegistry_Set_EmptyBackend(t *testing.T) {
	reg := NewDeadlineRegistry()
	if err := reg.Set("", time.Second); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestDeadlineRegistry_Set_NonPositiveDuration(t *testing.T) {
	reg := NewDeadlineRegistry()
	if err := reg.Set("http://backend:9000", 0); err == nil {
		t.Fatal("expected error for zero duration")
	}
	if err := reg.Set("http://backend:9000", -time.Second); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestDeadlineRegistry_Get_Missing(t *testing.T) {
	reg := NewDeadlineRegistry()
	_, ok := reg.Get("http://unknown:8080")
	if ok {
		t.Fatal("expected miss for unknown backend")
	}
}

func TestDeadlineRegistry_Delete(t *testing.T) {
	reg := NewDeadlineRegistry()
	_ = reg.Set("http://backend:9000", time.Second)
	reg.Delete("http://backend:9000")
	_, ok := reg.Get("http://backend:9000")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestDeadlineRegistry_Snapshot(t *testing.T) {
	reg := NewDeadlineRegistry()
	_ = reg.Set("http://a:8080", 1*time.Second)
	_ = reg.Set("http://b:8080", 2*time.Second)
	snap := reg.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	if snap["http://a:8080"] != 1*time.Second {
		t.Errorf("unexpected value for a: %v", snap["http://a:8080"])
	}
}
