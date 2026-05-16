package upstream

import (
	"net/http"
	"testing"
)

type mockHealthChecker struct {
	healthy map[string]bool
}

func (m *mockHealthChecker) IsHealthy(url string) bool {
	return m.healthy[url]
}

func newMockHealthChecker(healthy ...string) *mockHealthChecker {
	m := &mockHealthChecker{healthy: make(map[string]bool)}
	for _, h := range healthy {
		m.healthy[h] = true
	}
	return m
}

func TestNewPriorityBalancer_Empty(t *testing.T) {
	_, err := NewPriorityBalancer(nil, newMockHealthChecker())
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNewPriorityBalancer_NilHC(t *testing.T) {
	_, err := NewPriorityBalancer(map[string]int{"http://a": 1}, nil)
	if err == nil {
		t.Fatal("expected error for nil health checker")
	}
}

func TestPriorityBalancer_PicksHighestPriority(t *testing.T) {
	hc := newMockHealthChecker("http://a", "http://b", "http://c")
	pb, err := NewPriorityBalancer(map[string]int{
		"http://a": 1,
		"http://b": 10,
		"http://c": 5,
	}, hc)
	if err != nil {
		t.Fatal(err)
	}
	got := pb.Next(&http.Request{})
	if got != "http://b" {
		t.Errorf("expected http://b, got %s", got)
	}
}

func TestPriorityBalancer_SkipsUnhealthy(t *testing.T) {
	hc := newMockHealthChecker("http://c") // only c is healthy
	pb, err := NewPriorityBalancer(map[string]int{
		"http://a": 10,
		"http://b": 5,
		"http://c": 1,
	}, hc)
	if err != nil {
		t.Fatal(err)
	}
	got := pb.Next(&http.Request{})
	if got != "http://c" {
		t.Errorf("expected http://c, got %s", got)
	}
}

func TestPriorityBalancer_FallbackWhenNoneHealthy(t *testing.T) {
	hc := newMockHealthChecker() // none healthy
	pb, err := NewPriorityBalancer(map[string]int{
		"http://a": 10,
		"http://b": 1,
	}, hc)
	if err != nil {
		t.Fatal(err)
	}
	got := pb.Next(&http.Request{})
	if got == "" {
		t.Fatal("expected fallback backend, got empty string")
	}
}

func TestPriorityBalancer_Backends_ReturnsCopy(t *testing.T) {
	hc := newMockHealthChecker("http://a")
	pb, _ := NewPriorityBalancer(map[string]int{"http://a": 1}, hc)
	b1 := pb.Backends()
	b1[0] = "mutated"
	b2 := pb.Backends()
	if b2[0] == "mutated" {
		t.Error("Backends should return a copy")
	}
}
