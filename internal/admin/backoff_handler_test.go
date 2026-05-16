package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radovskyb/relayctl/internal/upstream"
)

func newTestBackoffRegistry(t *testing.T) *upstream.BackoffRegistry {
	t.Helper()
	return upstream.NewBackoffRegistry()
}

func TestBackoffHandler_Get_Empty(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	h := NewBackoffHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/backoff", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty map, got %v", out)
	}
}

func TestBackoffHandler_Put_Valid(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	h := NewBackoffHandler(reg)
	body, _ := json.Marshal(map[string]interface{}{
		"backend":      "http://svc:8080",
		"base_delay_ms": 100,
		"max_delay_ms":  5000,
		"multiplier":    2.0,
		"jitter":        0.1,
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/backoff", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	snap := reg.Snapshot()
	p, ok := snap["http://svc:8080"]
	if !ok {
		t.Fatal("expected entry in snapshot")
	}
	if p.BaseDelay != 100*time.Millisecond {
		t.Errorf("unexpected base delay: %v", p.BaseDelay)
	}
}

func TestBackoffHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	h := NewBackoffHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/backoff", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBackoffHandler_Put_MissingBackend(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	h := NewBackoffHandler(reg)
	body, _ := json.Marshal(map[string]interface{}{"base_delay_ms": 100})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/backoff", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBackoffHandler_Delete_Success(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	_ = reg.Set("http://svc:8080", upstream.BackoffPolicy{BaseDelay: time.Second, Multiplier: 2.0})
	h := NewBackoffHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/admin/backoff?backend=http://svc:8080", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Snapshot()["http://svc:8080"]; ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestBackoffHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestBackoffRegistry(t)
	h := NewBackoffHandler(reg)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/backoff", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
