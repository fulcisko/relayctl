package upstream

import (
	"testing"
)

func TestNewTagRegistry_Empty(t *testing.T) {
	r := NewTagRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	snap := r.Snapshot()
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestTagRegistry_Set_And_Get(t *testing.T) {
	r := NewTagRegistry()
	err := r.Set("http://backend:8080", []string{"primary", "us-east"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected backend to be found")
	}
	if len(tags) != 2 || tags[0] != "primary" || tags[1] != "us-east" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewTagRegistry()
	err := r.Set("", []string{"tag"})
	if err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestTagRegistry_Get_Missing(t *testing.T) {
	r := NewTagRegistry()
	_, ok := r.Get("http://missing:9000")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestTagRegistry_Delete(t *testing.T) {
	r := NewTagRegistry()
	_ = r.Set("http://backend:8080", []string{"canary"})
	r.Delete("http://backend:8080")
	_, ok := r.Get("http://backend:8080")
	if ok {
		t.Fatal("expected backend to be deleted")
	}
}

func TestTagRegistry_Snapshot_Isolation(t *testing.T) {
	r := NewTagRegistry()
	_ = r.Set("http://a:1", []string{"x"})
	_ = r.Set("http://b:2", []string{"y", "z"})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	// Mutate snapshot — should not affect registry
	snap["http://a:1"][0] = "mutated"
	tags, _ := r.Get("http://a:1")
	if tags[0] != "x" {
		t.Fatalf("registry was mutated via snapshot: %v", tags)
	}
}
