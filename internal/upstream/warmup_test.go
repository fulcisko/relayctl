package upstream

import (
	"testing"
	"time"
)

func TestNewWarmupRegistry_Empty(t *testing.T) {
	r := NewWarmupRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(r.entries))
	}
}

func TestWarmupRegistry_Register_EmptyBackend(t *testing.T) {
	r := NewWarmupRegistry()
	if err := r.Register("", WarmupConfig{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestWarmupRegistry_Weight_Unregistered(t *testing.T) {
	r := NewWarmupRegistry()
	if w := r.Weight("http://unknown"); w != 100 {
		t.Fatalf("expected 100 for unknown backend, got %d", w)
	}
}

func TestWarmupRegistry_Weight_FullyWarmed(t *testing.T) {
	r := NewWarmupRegistry()
	cfg := WarmupConfig{MaxWeight: 50, RampDuration: time.Millisecond}
	_ = r.Register("http://backend", cfg)
	time.Sleep(5 * time.Millisecond)
	if w := r.Weight("http://backend"); w != 50 {
		t.Fatalf("expected 50 after ramp, got %d", w)
	}
}

func TestWarmupRegistry_Weight_Ramping(t *testing.T) {
	r := NewWarmupRegistry()
	cfg := WarmupConfig{MaxWeight: 100, RampDuration: time.Hour}
	_ = r.Register("http://backend", cfg)
	w := r.Weight("http://backend")
	if w >= 100 {
		t.Fatalf("expected weight < 100 during ramp, got %d", w)
	}
}

func TestWarmupRegistry_Delete(t *testing.T) {
	r := NewWarmupRegistry()
	_ = r.Register("http://backend", WarmupConfig{})
	r.Delete("http://backend")
	if w := r.Weight("http://backend"); w != 100 {
		t.Fatalf("expected 100 after delete, got %d", w)
	}
}

func TestWarmupRegistry_Snapshot(t *testing.T) {
	r := NewWarmupRegistry()
	_ = r.Register("http://a", WarmupConfig{MaxWeight: 80, RampDuration: time.Hour})
	_ = r.Register("http://b", WarmupConfig{MaxWeight: 60, RampDuration: time.Millisecond})
	time.Sleep(5 * time.Millisecond)
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries in snapshot, got %d", len(snap))
	}
	if snap["http://b"] != 60 {
		t.Fatalf("expected b=60, got %d", snap["http://b"])
	}
}

func TestWarmupRegistry_DefaultsApplied(t *testing.T) {
	r := NewWarmupRegistry()
	_ = r.Register("http://x", WarmupConfig{MaxWeight: 0, RampDuration: 0})
	r.mu.RLock()
	e := r.entries["http://x"]
	r.mu.RUnlock()
	if e.config.MaxWeight != 100 {
		t.Fatalf("expected default MaxWeight=100, got %d", e.config.MaxWeight)
	}
	if e.config.RampDuration != 30*time.Second {
		t.Fatalf("expected default RampDuration=30s, got %v", e.config.RampDuration)
	}
}
