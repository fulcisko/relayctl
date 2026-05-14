package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHCFailover struct {
	healthy map[string]bool
}

func (m *mockHCFailover) IsHealthy(url string) bool {
	return m.healthy[url]
}

func TestNewFailoverBalancer_Empty(t *testing.T) {
	_, err := NewFailoverBalancer(nil, &mockHCFailover{})
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestFailoverBalancer_PrimaryHealthy(t *testing.T) {
	hc := &mockHCFailover{healthy: map[string]bool{"http://a": true, "http://b": true}}
	fb, _ := NewFailoverBalancer([]string{"http://a", "http://b"}, hc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := fb.Next(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://a" {
		t.Errorf("expected primary backend, got %s", got)
	}
}

func TestFailoverBalancer_FallsBackToSecondary(t *testing.T) {
	hc := &mockHCFailover{healthy: map[string]bool{"http://a": false, "http://b": true}}
	fb, _ := NewFailoverBalancer([]string{"http://a", "http://b"}, hc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := fb.Next(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://b" {
		t.Errorf("expected fallback backend, got %s", got)
	}
}

func TestFailoverBalancer_NoneHealthy(t *testing.T) {
	hc := &mockHCFailover{healthy: map[string]bool{}}
	fb, _ := NewFailoverBalancer([]string{"http://a", "http://b"}, hc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := fb.Next(req)
	if err != ErrNoHealthyBackend {
		t.Errorf("expected ErrNoHealthyBackend, got %v", err)
	}
}

func TestFailoverBalancer_Update(t *testing.T) {
	hc := &mockHCFailover{healthy: map[string]bool{"http://c": true}}
	fb, _ := NewFailoverBalancer([]string{"http://a"}, hc)
	if err := fb.Update([]string{"http://c"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, _ := fb.Next(req)
	if got != "http://c" {
		t.Errorf("expected http://c, got %s", got)
	}
}

func TestFailoverBalancer_Update_Empty(t *testing.T) {
	hc := &mockHCFailover{}
	fb, _ := NewFailoverBalancer([]string{"http://a"}, hc)
	if err := fb.Update(nil); err == nil {
		t.Fatal("expected error for empty update")
	}
}

func TestFailoverBalancer_Backends_ReturnsCopy(t *testing.T) {
	hc := &mockHCFailover{}
	fb, _ := NewFailoverBalancer([]string{"http://a", "http://b"}, hc)
	bs := fb.Backends()
	bs[0] = "mutated"
	if fb.backends[0] == "mutated" {
		t.Error("Backends() should return a copy")
	}
}
