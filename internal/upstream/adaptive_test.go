package upstream

import (
	"testing"
	"time"
)

func TestNewAdaptiveBalancer_Empty(t *testing.T) {
	_, err := NewAdaptiveBalancer([]string{})
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNewAdaptiveBalancer_Valid(t *testing.T) {
	b, err := NewAdaptiveBalancer([]string{"a:80", "b:80"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Backends()) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(b.Backends()))
	}
}

func TestAdaptiveBalancer_Next_ReturnsMember(t *testing.T) {
	backends := []string{"a:80", "b:80", "c:80"}
	b, _ := NewAdaptiveBalancer(backends)

	got, done := b.Next("")
	done()

	found := false
	for _, addr := range backends {
		if addr == got {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Next() returned unknown backend %q", got)
	}
}

func TestAdaptiveBalancer_Next_PrefersLowerLatency(t *testing.T) {
	b, _ := NewAdaptiveBalancer([]string{"fast:80", "slow:80"})

	// Seed latency: slow backend gets a high latency recorded.
	b.mu.Lock()
	b.latency["slow:80"] = 500 * time.Millisecond
	b.samples["slow:80"] = 1
	b.latency["fast:80"] = 10 * time.Millisecond
	b.samples["fast:80"] = 1
	b.mu.Unlock()

	got, done := b.Next("")
	done()

	if got != "fast:80" {
		t.Errorf("expected fast:80, got %q", got)
	}
}

func TestAdaptiveBalancer_Done_UpdatesLatency(t *testing.T) {
	b, _ := NewAdaptiveBalancer([]string{"x:80"})

	_, done := b.Next("")
	time.Sleep(5 * time.Millisecond)
	done()

	latency := b.AvgLatency("x:80")
	if latency <= 0 {
		t.Errorf("expected positive latency, got %v", latency)
	}
}

func TestAdaptiveBalancer_Backends_ReturnsCopy(t *testing.T) {
	b, _ := NewAdaptiveBalancer([]string{"a:80"})
	copy1 := b.Backends()
	copy1[0] = "mutated"
	copy2 := b.Backends()
	if copy2[0] == "mutated" {
		t.Error("Backends() returned internal slice reference")
	}
}

func TestAdaptiveBalancer_EMA_Convergence(t *testing.T) {
	b, _ := NewAdaptiveBalancer([]string{"a:80"})

	// Simulate multiple done() calls to verify EMA updates.
	for i := 0; i < 5; i++ {
		_, done := b.Next("")
		done()
	}

	if b.samples["a:80"] != 5 {
		t.Errorf("expected 5 samples, got %d", b.samples["a:80"])
	}
}

func TestAdaptiveBalancer_AvgLatency_UnknownBackend(t *testing.T) {
	b, _ := NewAdaptiveBalancer([]string{"a:80"})

	// Querying a backend that has never been used should return zero.
	latency := b.AvgLatency("unknown:80")
	if latency != 0 {
		t.Errorf("expected zero latency for unknown backend, got %v", latency)
	}
}
