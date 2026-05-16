package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

func newTestPassthroughRegistry() *upstream.PassthroughRegistry {
	return upstream.NewPassthroughRegistry()
}

func TestPassthroughHandler_Get_Empty(t *testing.T) {
	h := NewPassthroughHandler(newTestPassthroughRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/passthrough", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp["backends"]) != 0 {
		t.Errorf("expected empty list, got %v", resp["backends"])
	}
}

func TestPassthroughHandler_Put_Valid(t *testing.T) {
	reg := newTestPassthroughRegistry()
	h := NewPassthroughHandler(reg)
	body := bytes.NewBufferString(`{"backend":"http://svc:9000"}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/passthrough", body))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if !reg.IsPassthrough("http://svc:9000") {
		t.Error("expected backend to be registered")
	}
}

func TestPassthroughHandler_Put_InvalidJSON(t *testing.T) {
	h := NewPassthroughHandler(newTestPassthroughRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/passthrough", bytes.NewBufferString(`not-json`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPassthroughHandler_Put_EmptyBackend(t *testing.T) {
	h := NewPassthroughHandler(newTestPassthroughRegistry())
	body := bytes.NewBufferString(`{"backend":""}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/passthrough", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPassthroughHandler_Delete_Success(t *testing.T) {
	reg := newTestPassthroughRegistry()
	_ = reg.Set("http://svc:9000")
	h := NewPassthroughHandler(reg)
	body := bytes.NewBufferString(`{"backend":"http://svc:9000"}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/admin/passthrough", body))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if reg.IsPassthrough("http://svc:9000") {
		t.Error("expected backend to be removed")
	}
}

func TestPassthroughHandler_MethodNotAllowed(t *testing.T) {
	h := NewPassthroughHandler(newTestPassthroughRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/passthrough", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
