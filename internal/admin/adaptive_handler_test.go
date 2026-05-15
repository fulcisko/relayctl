package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/relayctl/internal/upstream"
)

func newTestAdaptiveBalancer(t *testing.T, backends []string) *upstream.AdaptiveBalancer {
	t.Helper()
	b, err := upstream.NewAdaptiveBalancer(backends)
	if err != nil {
		t.Fatalf("NewAdaptiveBalancer: %v", err)
	}
	return b
}

func TestAdaptiveHandler_OK(t *testing.T) {
	b := newTestAdaptiveBalancer(t, []string{"a:80", "b:80"})
	h := NewAdaptiveHandler(b)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/adaptive", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var snaps []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&snaps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(snaps) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snaps))
	}
}

func TestAdaptiveHandler_MethodNotAllowed(t *testing.T) {
	b := newTestAdaptiveBalancer(t, []string{"a:80"})
	h := NewAdaptiveHandler(b)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/adaptive", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAdaptiveHandler_NilBalancer(t *testing.T) {
	h := NewAdaptiveHandler(nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/adaptive", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "[]" {
		t.Errorf("expected empty array, got %q", body)
	}
}

func TestAdaptiveHandler_ReflectsLatency(t *testing.T) {
	b := newTestAdaptiveBalancer(t, []string{"x:80"})

	// Trigger a request cycle so latency is recorded.
	_, done := b.Next("")
	time.Sleep(2 * time.Millisecond)
	done()

	h := NewAdaptiveHandler(b)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/adaptive", nil)
	h.ServeHTTP(rec, req)

	var snaps []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&snaps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snaps))
	}
	latency, ok := snaps[0]["avg_latency_ns"].(float64)
	if !ok || latency <= 0 {
		t.Errorf("expected positive avg_latency_ns, got %v", snaps[0]["avg_latency_ns"])
	}
}
