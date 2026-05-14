package upstream

import (
	"math/rand"
	"testing"
)

func TestNewRandomBalancer_Empty(t *testing.T) {
	_, err := NewRandomBalancer(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty backend list")
	}
}

func TestNewRandomBalancer_Valid(t *testing.T) {
	b, err := NewRandomBalancer([]string{"http://a", "http://b"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Backends()) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(b.Backends()))
	}
}

func TestRandomBalancer_Next_ReturnsMember(t *testing.T) {
	backends := []string{"http://a", "http://b", "http://c"}
	b, _ := NewRandomBalancer(backends, rand.NewSource(42))
	seen := map[string]bool{}
	for i := 0; i < 30; i++ {
		got, err := b.Next("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[got] = true
	}
	for _, be := range backends {
		if !seen[be] {
			t.Errorf("backend %q never selected in 30 draws", be)
		}
	}
}

func TestRandomBalancer_Next_SingleBackend(t *testing.T) {
	b, _ := NewRandomBalancer([]string{"http://only"}, nil)
	got, err := b.Next("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://only" {
		t.Errorf("expected http://only, got %s", got)
	}
}

func TestRandomBalancer_Backends_ReturnsCopy(t *testing.T) {
	b, _ := NewRandomBalancer([]string{"http://a"}, nil)
	copy1 := b.Backends()
	copy1[0] = "mutated"
	copy2 := b.Backends()
	if copy2[0] == "mutated" {
		t.Error("Backends() returned internal slice reference")
	}
}

func TestRandomBalancer_Update_Valid(t *testing.T) {
	b, _ := NewRandomBalancer([]string{"http://old"}, nil)
	if err := b.Update([]string{"http://new1", "http://new2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Backends()) != 2 {
		t.Errorf("expected 2 backends after update, got %d", len(b.Backends()))
	}
}

func TestRandomBalancer_Update_Empty(t *testing.T) {
	b, _ := NewRandomBalancer([]string{"http://a"}, nil)
	if err := b.Update([]string{}); err == nil {
		t.Error("expected error when updating with empty list")
	}
}
