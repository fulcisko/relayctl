package upstream

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func newSimpleBurstBalancer(backends []string) *roundRobinBalancer {
	b, _ := New(backends)
	return b
}

func TestNewBurstBalancer_NilPrimary(t *testing.T) {
	_, err := NewBurstBalancer(nil, newSimpleBurstBalancer([]string{"b:80"}), 10, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewBurstBalancer_NilBurst(t *testing.T) {
	_, err := NewBurstBalancer(newSimpleBurstBalancer([]string{"a:80"}), nil, 10, time.Second, time.Second)
	if err == nil {
		t.Fatal("expected error for nil burst")
	}
}

func TestBurstBalancer_DefaultsOnZeroValues(t *testing.T) {
	primary, _ := New([]string{"p:80"})
	burst, _ := New([]string{"b:80"})
	bb, err := NewBurstBalancer(primary, burst, 0, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bb.threshold != 100 {
		t.Errorf("expected default threshold 100, got %d", bb.threshold)
	}
}

func TestBurstBalancer_NoBurstBelowThreshold(t *testing.T) {
	primary, _ := New([]string{"primary:80"})
	burst, _ := New([]string{"burst:80"})
	bb, _ := NewBurstBalancer(primary, burst, 10, time.Second, 5*time.Second)

	req := Request{HTTPRequest: &http.Request{}}
	for i := 0; i < 5; i++ {
		backend, err := bb.Next(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if backend != "primary:80" {
			t.Errorf("expected primary:80, got %s", backend)
		}
	}
	if bb.IsBurstActive() {
		t.Error("burst should not be active below threshold")
	}
}

func TestBurstBalancer_ActivatesAtThreshold(t *testing.T) {
	primary, _ := New([]string{"primary:80"})
	burst, _ := New([]string{"burst:80"})
	bb, _ := NewBurstBalancer(primary, burst, 5, time.Second, 10*time.Second)

	req := Request{HTTPRequest: &http.Request{}}
	for i := 0; i < 5; i++ {
		bb.Next(req) //nolint
	}
	if !bb.IsBurstActive() {
		t.Error("burst should be active at threshold")
	}
	backend, err := bb.Next(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend != "burst:80" {
		t.Errorf("expected burst:80 during burst mode, got %s", backend)
	}
}

func TestBurstBalancer_ExpiresBurstTTL(t *testing.T) {
	primary, _ := New([]string{"primary:80"})
	burst, _ := New([]string{"burst:80"})
	bb, _ := NewBurstBalancer(primary, burst, 2, time.Second, 50*time.Millisecond)

	req := Request{HTTPRequest: &http.Request{}}
	bb.Next(req) //nolint
	bb.Next(req) //nolint
	if !bb.IsBurstActive() {
		t.Fatal("expected burst to be active")
	}
	time.Sleep(60 * time.Millisecond)
	if bb.IsBurstActive() {
		t.Error("burst should have expired after TTL")
	}
}

func TestBurstBalancer_Backends_Union(t *testing.T) {
	primary, _ := New([]string{"p1:80", "p2:80"})
	burst, _ := New([]string{"b1:80"})
	bb, _ := NewBurstBalancer(primary, burst, 10, time.Second, time.Second)

	backends := bb.Backends()
	if len(backends) != 3 {
		t.Errorf("expected 3 backends, got %d: %v", len(backends), backends)
	}
}

func TestBurstBalancer_WindowReset(t *testing.T) {
	primary, _ := New([]string{"primary:80"})
	burst, _ := New([]string{"burst:80"})
	bb, _ := NewBurstBalancer(primary, burst, 5, 50*time.Millisecond, 10*time.Second)

	req := Request{HTTPRequest: &http.Request{}}
	for i := 0; i < 4; i++ {
		bb.Next(req) //nolint
	}
	time.Sleep(60 * time.Millisecond)
	// window resets — count drops back to 0
	bb.mu.Lock()
	oldCount := bb.count
	bb.mu.Unlock()
	bb.Next(req) //nolint
	bb.mu.Lock()
	newCount := bb.count
	bb.mu.Unlock()
	if newCount >= oldCount && oldCount > 1 {
		t.Errorf("expected count to reset, old=%d new=%d", oldCount, newCount)
	}
}

// Ensure BurstBalancer satisfies the Balancer interface at compile time.
var _ Balancer = (*BurstBalancer)(nil)

func init() {
	// suppress unused import
	_ = fmt.Sprintf
}
