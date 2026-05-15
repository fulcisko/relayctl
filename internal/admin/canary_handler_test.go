package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/relayctl/internal/upstream"
)

func newTestCanary(t *testing.T, pct int) *upstream.CanaryBalancer {
	t.Helper()
	s, err := upstream.New([]string{"http://stable:8000"})
	if err != nil {
		t.Fatalf("stable balancer: %v", err)
	}
	c, err := upstream.New([]string{"http://canary:9000"})
	if err != nil {
		t.Fatalf("canary balancer: %v", err)
	}
	cb, err := upstream.NewCanaryBalancer(s, c, pct)
	if err != nil {
		t.Fatalf("canary: %v", err)
	}
	return cb
}

func TestCanaryHandler_Get(t *testing.T) {
	cb := newTestCanary(t, 20)
	h := NewCanaryHandler(cb)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/canary", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var snap struct {
		Percent  int      `json:"percent"`
		Backends []string `json:"backends"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Percent != 20 {
		t.Errorf("expected percent 20, got %d", snap.Percent)
	}
	if len(snap.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(snap.Backends))
	}
}

func TestCanaryHandler_Put_Valid(t *testing.T) {
	cb := newTestCanary(t, 10)
	h := NewCanaryHandler(cb)

	body, _ := json.Marshal(map[string]int{"percent": 40})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPut, "/admin/canary", bytes.NewReader(body)))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if cb.Percent() != 40 {
		t.Errorf("expected percent 40, got %d", cb.Percent())
	}
}

func TestCanaryHandler_Put_InvalidJSON(t *testing.T) {
	cb := newTestCanary(t, 10)
	h := NewCanaryHandler(cb)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPut, "/admin/canary", bytes.NewBufferString("not-json")))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCanaryHandler_MethodNotAllowed(t *testing.T) {
	cb := newTestCanary(t, 10)
	h := NewCanaryHandler(cb)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/admin/canary", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestCanaryHandler_NilBalancer(t *testing.T) {
	h := NewCanaryHandler(nil)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/canary", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}
