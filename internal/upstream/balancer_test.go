package upstream

import (
	"testing"
)

func TestNew_EmptyBackends(t *testing.T) {
	_, err := New([]string{})
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNew_ValidBackends(t *testing.T) {
	b, err := New([]string{"http://localhost:8081"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil balancer")
	}
}

func TestNext_RoundRobin(t *testing.T) {
	backends := []string{"http://a", "http://b", "http://c"}
	b, _ := New(backends)

	for i, want := range []string{"http://a", "http://b", "http://c", "http://a"} {
		got, err := b.Next()
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != want {
			t.Errorf("call %d: got %q, want %q", i, got, want)
		}
	}
}

func TestNext_SingleBackend(t *testing.T) {
	b, _ := New([]string{"http://only"})
	for i := 0; i < 5; i++ {
		got, err := b.Next()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://only" {
			t.Errorf("expected http://only, got %q", got)
		}
	}
}

func TestBackends_ReturnsCopy(t *testing.T) {
	original := []string{"http://a", "http://b"}
	b, _ := New(original)
	copy := b.Backends()
	copy[0] = "http://mutated"
	if b.Backends()[0] == "http://mutated" {
		t.Error("Backends() should return a copy, not a reference")
	}
}

func TestUpdate_ReplacesBackends(t *testing.T) {
	b, _ := New([]string{"http://old"})
	err := b.Update([]string{"http://new1", "http://new2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := b.Next()
	if got != "http://new1" {
		t.Errorf("expected http://new1, got %q", got)
	}
}

func TestUpdate_EmptyBackends(t *testing.T) {
	b, _ := New([]string{"http://a"})
	err := b.Update([]string{})
	if err == nil {
		t.Error("expected error when updating with empty backends")
	}
}
