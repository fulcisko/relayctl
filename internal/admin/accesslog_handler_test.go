package admin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/relayctl/internal/admin"
)

func newTestCollector(t *testing.T, entries int) *admin.AccessLogCollector {
	t.Helper()
	c := admin.NewAccessLogCollector(50)
	for i := 0; i < entries; i++ {
		line := fmt.Sprintf(`{"method":"GET","path":"/p%d","status":200}\n`, i)
		_, _ = c.Write([]byte(line))
	}
	return c
}

func TestAccessLogHandler_OK(t *testing.T) {
	c := newTestCollector(t, 3)
	h := admin.NewAccessLogHandler(c)

	req := httptest.NewRequest(http.MethodGet, "/admin/accesslog", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	count, ok := body["count"].(float64)
	if !ok || int(count) != 3 {
		t.Errorf("expected count=3, got %v", body["count"])
	}
}

func TestAccessLogHandler_MethodNotAllowed(t *testing.T) {
	c := admin.NewAccessLogCollector(10)
	h := admin.NewAccessLogHandler(c)

	req := httptest.NewRequest(http.MethodPost, "/admin/accesslog", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAccessLogHandler_EmptyCollector(t *testing.T) {
	c := admin.NewAccessLogCollector(10)
	h := admin.NewAccessLogHandler(c)

	req := httptest.NewRequest(http.MethodGet, "/admin/accesslog", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	count := body["count"].(float64)
	if int(count) != 0 {
		t.Errorf("expected count=0, got %v", count)
	}
}

func TestAccessLogCollector_RingBuffer(t *testing.T) {
	c := admin.NewAccessLogCollector(3)
	for i := 0; i < 5; i++ {
		line := fmt.Sprintf(`{"status":%d}`, 200+i)
		_, _ = c.Write([]byte(line))
	}
	snap := c.Snapshot()
	if len(snap) != 3 {
		t.Errorf("ring buffer: expected 3 entries, got %d", len(snap))
	}
}
