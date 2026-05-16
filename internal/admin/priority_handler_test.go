package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukethinnes/relayctl/internal/upstream"
)

type mockPriorityHealthChecker struct {
	healthy map[string]bool
}

func (m *mockPriorityHealthChecker) IsHealthy(backend string) bool {
	v, ok := m.healthy[backend]
	return ok && v
}

func newTestPriorityBalancer() *upstream.PriorityBalancer {
	groups := []upstream.PriorityGroup{
		{Backend: "http://primary:8080", Priority: 10},
		{Backend: "http://secondary:8080", Priority: 5},
	}
	hc := &mockPriorityHealthChecker{
		healthy: map[string]bool{
			"http://primary:8080":   true,
			"http://secondary:8080": true,
		},
	}
	return upstream.NewPriorityBalancer(groups, hc)
}

func TestPriorityHandler_Get(t *testing.T) {
	b := newTestPriorityBalancer()
	h := NewPriorityHandler(b)

	req := httptest.NewRequest(http.MethodGet, "/admin/priority", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Groups []upstream.PriorityGroup `json:"groups"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(resp.Groups))
	}
}

func TestPriorityHandler_MethodNotAllowed(t *testing.T) {
	b := newTestPriorityBalancer()
	h := NewPriorityHandler(b)

	req := httptest.NewRequest(http.MethodPost, "/admin/priority", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestPriorityHandler_NilBalancer(t *testing.T) {
	h := NewPriorityHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/priority", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestPriorityHandler_ContentType(t *testing.T) {
	b := newTestPriorityBalancer()
	h := NewPriorityHandler(b)

	req := httptest.NewRequest(http.MethodGet, "/admin/priority", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}
