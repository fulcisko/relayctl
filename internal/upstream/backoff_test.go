package upstream

import (
	"testing"
	"time"
)

func TestNewBackoffRegistry_Empty(t *testing.T) {
	r := NewBackoffRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestBackoffRegistry_Set_And_Next(t *testing.T) {
	r := NewBackoffRegistry()
	err := r.Set("http://backend", BackoffPolicy{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d1, ok := r.Next("http://backend")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	d2, _ := r.Next("http://backend")
	if d2 <= d1 {
		t.Errorf("expected delay to grow: d1=%v d2=%v", d1, d2)
	}
}

func TestBackoffRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewBackoffRegistry()
	if err := r.Set("", BackoffPolicy{BaseDelay: time.Second}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestBackoffRegistry_Next_Missing(t *testing.T) {
	r := NewBackoffRegistry()
	_, ok := r.Next("http://unknown")
	if ok {
		t.Fatal("expected false for unknown backend")
	}
}

func TestBackoffRegistry_Reset(t *testing.T) {
	r := NewBackoffRegistry()
	_ = r.Set("http://backend", BackoffPolicy{
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
	})
	r.Next("http://backend")
	r.Next("http://backend")
	r.Reset("http://backend")
	d, _ := r.Next("http://backend")
	if d > 100*time.Millisecond {
		t.Errorf("expected reset to base delay, got %v", d)
	}
}

func TestBackoffRegistry_Delete(t *testing.T) {
	r := NewBackoffRegistry()
	_ = r.Set("http://backend", BackoffPolicy{BaseDelay: time.Second, Multiplier: 2.0})
	r.Delete("http://backend")
	if _, ok := r.Next("http://backend"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestBackoffRegistry_MaxDelay_Capped(t *testing.T) {
	r := NewBackoffRegistry()
	_ = r.Set("http://backend", BackoffPolicy{
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   600 * time.Millisecond,
		Multiplier: 10.0,
	})
	for i := 0; i < 5; i++ {
		r.Next("http://backend")
	}
	d, _ := r.Next("http://backend")
	if d > 600*time.Millisecond {
		t.Errorf("expected delay capped at MaxDelay, got %v", d)
	}
}
