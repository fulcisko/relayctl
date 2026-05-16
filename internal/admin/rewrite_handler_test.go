package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/internal/upstream"
)

func newTestRewriteRegistry() *upstream.RewriteRegistry {
	return upstream.NewRewriteRegistry()
}

func TestRewriteHandler_Get_Empty(t *testing.T) {
	h := NewRewriteHandler(newTestRewriteRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/rewrite", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestRewriteHandler_Put_Valid(t *testing.T) {
	reg := newTestRewriteRegistry()
	h := NewRewriteHandler(reg)
	body := `{"backend":"http://svc:8080","prefix":"/api","replacement":"/"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/rewrite", bytes.NewBufferString(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("http://svc:8080"); !ok {
		t.Fatal("expected rule to be stored")
	}
}

func TestRewriteHandler_Put_InvalidJSON(t *testing.T) {
	h := NewRewriteHandler(newTestRewriteRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/rewrite", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRewriteHandler_Put_MissingBackend(t *testing.T) {
	h := NewRewriteHandler(newTestRewriteRegistry())
	body := `{"backend":"","prefix":"/api","replacement":""}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/rewrite", bytes.NewBufferString(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRewriteHandler_Delete_Success(t *testing.T) {
	reg := newTestRewriteRegistry()
	_ = reg.Set("http://svc:9000", "/old", "/new")
	h := NewRewriteHandler(reg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/rewrite?backend=http://svc:9000", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("http://svc:9000"); ok {
		t.Fatal("expected rule to be deleted")
	}
}

func TestRewriteHandler_Delete_MissingParam(t *testing.T) {
	h := NewRewriteHandler(newTestRewriteRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/admin/rewrite", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRewriteHandler_MethodNotAllowed(t *testing.T) {
	h := NewRewriteHandler(newTestRewriteRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/rewrite", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRewriteHandler_NilRegistry(t *testing.T) {
	h := NewRewriteHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/rewrite", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
