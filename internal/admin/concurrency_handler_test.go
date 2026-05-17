package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasmalkmus/relayctl/internal/upstream"
)

func newTestConcurrencyRegistry() *upstream.ConcurrencyRegistry {
	return upstream.NewConcurrencyRegistry()
}

func TestConcurrencyHandler_Get_NotFound(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	req := httptest.NewRequest(http.MethodGet, "/admin/concurrency?backend=http://x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestConcurrencyHandler_Get_MissingParam(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	req := httptest.NewRequest(http.MethodGet, "/admin/concurrency", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConcurrencyHandler_Put_Valid(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	body, _ := json.Marshal(map[string]interface{}{"backend": "http://b:9000", "max": 50})
	req := httptest.NewRequest(http.MethodPut, "/admin/concurrency", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	max, _, ok := reg.Get("http://b:9000")
	if !ok || max != 50 {
		t.Fatalf("expected max=50, got %d ok=%v", max, ok)
	}
}

func TestConcurrencyHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	req := httptest.NewRequest(http.MethodPut, "/admin/concurrency", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConcurrencyHandler_Put_InvalidMax(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	body, _ := json.Marshal(map[string]interface{}{"backend": "http://b", "max": 0})
	req := httptest.NewRequest(http.MethodPut, "/admin/concurrency", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConcurrencyHandler_Delete_Success(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	_ = reg.Set("http://b", 10)
	h := NewConcurrencyHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/concurrency?backend=http://b", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, _, ok := reg.Get("http://b")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestConcurrencyHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestConcurrencyRegistry()
	h := NewConcurrencyHandler(reg)
	req := httptest.NewRequest(http.MethodPost, "/admin/concurrency", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
