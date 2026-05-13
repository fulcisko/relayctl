package upstream

import (
	"testing"
)

// mockHealthChecker implements HealthChecker for testing.
type mockHealthChecker struct {
	healthy map[string]bool
}

func (m *mockHealthChecker) IsHealthy(url string) bool {
	return m.healthy[url]
}

func newMockHC(healthy ...string) *mockHealthChecker {
	m := &mockHealthChecker{healthy: make(map[string]bool)}
	for _, h := range healthy {
		m.healthy[h] = true
	}
	return m
}

func TestFilteredBalancer_Next_AllHealthy(t *testing.T) {
	bal, _ := New([]string{"http://a:8080", "http://b:8080"})
	hc := newMockHC("http://a:8080", "http://b:8080")
	fb := NewFilteredBalancer(bal, hc)

	got := fb.Next()
	if got == "" {
		t.Fatal("expected a backend, got empty string")
	}
}

func TestFilteredBalancer_Next_SkipsUnhealthy(t *testing.T) {
	bal, _ := New([]string{"http://a:8080", "http://b:8080"})
	hc := newMockHC("http://b:8080") // only b is healthy
	fb := NewFilteredBalancer(bal, hc)

	for i := 0; i < 5; i++ {
		got := fb.Next()
		if got != "http://b:8080" {
			t.Fatalf("expected http://b:8080, got %q", got)
		}
	}
}

func TestFilteredBalancer_Next_NoneHealthy(t *testing.T) {
	bal, _ := New([]string{"http://a:8080", "http://b:8080"})
	hc := newMockHC() // none healthy
	fb := NewFilteredBalancer(bal, hc)

	got := fb.Next()
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestFilteredBalancer_HealthyBackends(t *testing.T) {
	bal, _ := New([]string{"http://a:8080", "http://b:8080", "http://c:8080"})
	hc := newMockHC("http://a:8080", "http://c:8080")
	fb := NewFilteredBalancer(bal, hc)

	healthy := fb.HealthyBackends()
	if len(healthy) != 2 {
		t.Fatalf("expected 2 healthy backends, got %d", len(healthy))
	}
}

func TestFilteredBalancer_UpdateBalancer(t *testing.T) {
	bal, _ := New([]string{"http://old:8080"})
	hc := newMockHC("http://new:8080")
	fb := NewFilteredBalancer(bal, hc)

	newBal, _ := New([]string{"http://new:8080"})
	fb.UpdateBalancer(newBal)

	got := fb.Next()
	if got != "http://new:8080" {
		t.Fatalf("expected http://new:8080, got %q", got)
	}
}
