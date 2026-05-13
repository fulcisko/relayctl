package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

type stubHealthChecker struct {
	healthy map[string]bool
}

func (s *stubHealthChecker) IsHealthy(url string) bool {
	return s.healthy[url]
}

func newTestFilteredBalancer(backends []string, healthyOnes []string) *upstream.FilteredBalancer {
	bal, _ := upstream.New(backends)
	hc := &stubHealthChecker{healthy: make(map[string]bool)}
	for _, h := range healthyOnes {
		hc.healthy[h] = true
	}
	return upstream.NewFilteredBalancer(bal, hc)
}

func TestHealthFilterHandler_OK(t *testing.T) {
	fb := newTestFilteredBalancer(
		[]string{"http://a:8080", "http://b:8080"},
		[]string{"http://a:8080"},
	)
	h := NewHealthFilterHandler(fb)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/healthy-backends", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Healthy []string `json:"healthy"`
		Count   int      `json:"count"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Count != 1 {
		t.Fatalf("expected count 1, got %d", resp.Count)
	}
	if resp.Healthy[0] != "http://a:8080" {
		t.Fatalf("unexpected backend: %s", resp.Healthy[0])
	}
}

func TestHealthFilterHandler_MethodNotAllowed(t *testing.T) {
	fb := newTestFilteredBalancer([]string{"http://a:8080"}, []string{"http://a:8080"})
	h := NewHealthFilterHandler(fb)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/healthy-backends", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHealthFilterHandler_NilBalancer(t *testing.T) {
	h := NewHealthFilterHandler(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/healthy-backends", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Count != 0 {
		t.Fatalf("expected count 0, got %d", resp.Count)
	}
}
