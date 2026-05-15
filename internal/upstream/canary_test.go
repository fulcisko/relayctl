package upstream

import (
	"net/http"
	"testing"
)

func newSimpleBalancer(backends []string) Balancer {
	b, _ := New(backends)
	return b
}

func TestNewCanaryBalancer_NilStable(t *testing.T) {
	c, _ := newSimpleBalancer([]string{"http://canary:9000"}).(*RoundRobinBalancer)
	_, err := NewCanaryBalancer(nil, c, 10)
	if err == nil {
		t.Fatal("expected error for nil stable balancer")
	}
}

func TestNewCanaryBalancer_NilCanary(t *testing.T) {
	s := newSimpleBalancer([]string{"http://stable:8000"})
	_, err := NewCanaryBalancer(s, nil, 10)
	if err == nil {
		t.Fatal("expected error for nil canary balancer")
	}
}

func TestNewCanaryBalancer_ClampsPercent(t *testing.T) {
	s := newSimpleBalancer([]string{"http://stable:8000"})
	can := newSimpleBalancer([]string{"http://canary:9000"})

	cb, err := NewCanaryBalancer(s, can, 150)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.Percent() != 100 {
		t.Errorf("expected percent clamped to 100, got %d", cb.Percent())
	}
}

func TestCanaryBalancer_ZeroPercent_AlwaysStable(t *testing.T) {
	stableBackend := "http://stable:8000"
	canaryBackend := "http://canary:9000"
	s := newSimpleBalancer([]string{stableBackend})
	can := newSimpleBalancer([]string{canaryBackend})

	cb, _ := NewCanaryBalancer(s, can, 0)
	req := Request{HTTPRequest: &http.Request{}}

	for i := 0; i < 50; i++ {
		backend, done, err := cb.Next(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if done != nil {
			done()
		}
		if backend != stableBackend {
			t.Errorf("expected stable backend, got %s", backend)
		}
	}
}

func TestCanaryBalancer_FullPercent_AlwaysCanary(t *testing.T) {
	stableBackend := "http://stable:8000"
	canaryBackend := "http://canary:9000"
	s := newSimpleBalancer([]string{stableBackend})
	can := newSimpleBalancer([]string{canaryBackend})

	cb, _ := NewCanaryBalancer(s, can, 100)
	req := Request{HTTPRequest: &http.Request{}}

	for i := 0; i < 50; i++ {
		backend, done, err := cb.Next(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if done != nil {
			done()
		}
		if backend != canaryBackend {
			t.Errorf("expected canary backend, got %s", backend)
		}
	}
}

func TestCanaryBalancer_Backends_ReturnsAll(t *testing.T) {
	s := newSimpleBalancer([]string{"http://stable:8000"})
	can := newSimpleBalancer([]string{"http://canary:9000"})
	cb, _ := NewCanaryBalancer(s, can, 50)

	backends := cb.Backends()
	if len(backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(backends))
	}
}

func TestCanaryBalancer_SetPercent(t *testing.T) {
	s := newSimpleBalancer([]string{"http://stable:8000"})
	can := newSimpleBalancer([]string{"http://canary:9000"})
	cb, _ := NewCanaryBalancer(s, can, 10)

	cb.SetPercent(75)
	if cb.Percent() != 75 {
		t.Errorf("expected 75, got %d", cb.Percent())
	}

	cb.SetPercent(-5)
	if cb.Percent() != 0 {
		t.Errorf("expected 0 after negative clamp, got %d", cb.Percent())
	}
}
