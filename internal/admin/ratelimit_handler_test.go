package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/relayctl/internal/ratelimit"
)

func newTestRateLimitHandler() *RateLimitHandler {
	l := ratelimit.New(100, 60*time.Second)
	cfg := RateLimitConfig{Max: 100, Window: 60 * time.Second}
	return NewRateLimitHandler(l, cfg)
}

func TestRateLimitHandler_OK(t *testing.T) {
	h := newTestRateLimitHandler()
	req := httptest.NewRequest(http.MethodGet, "/admin/ratelimit", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rw.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if v, ok := body["max"]; !ok || v.(float64) != 100 {
		t.Errorf("expected max=100, got %v", body["max"])
	}
	if v, ok := body["window_seconds"]; !ok || v.(float64) != 60 {
		t.Errorf("expected window_seconds=60, got %v", body["window_seconds"])
	}
}

func TestRateLimitHandler_MethodNotAllowed(t *testing.T) {
	h := newTestRateLimitHandler()
	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		req := httptest.NewRequest(method, "/admin/ratelimit", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		if rw.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rw.Code)
		}
	}
}

func TestRateLimitHandler_ContentType(t *testing.T) {
	h := newTestRateLimitHandler()
	req := httptest.NewRequest(http.MethodGet, "/admin/ratelimit", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	ct := rw.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
