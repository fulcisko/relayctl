package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

func newTestResponseHeaderRegistry() *upstream.ResponseHeaderRegistry {
	return upstream.NewResponseHeaderRegistry()
}

func TestResponseHeaderHandler_Get_Empty(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	h := NewResponseHeaderHandler(reg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/response-headers", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
}

func TestResponseHeaderHandler_Put_Valid(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	h := NewResponseHeaderHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"backend": "http://backend:9000",
		"rules": map[string]interface{}{
			"set":    map[string]string{"X-Powered-By": "relayctl"},
			"delete": []string{"Server"},
		},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/response-headers", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	rules, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected rules to be stored")
	}
	if rules.Set["X-Powered-By"] != "relayctl" {
		t.Errorf("unexpected Set value: %v", rules.Set)
	}
}

func TestResponseHeaderHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	h := NewResponseHeaderHandler(reg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/response-headers", bytes.NewBufferString("not-json"))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestResponseHeaderHandler_Put_MissingBackend(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	h := NewResponseHeaderHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"backend": "",
		"rules":   map[string]interface{}{},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/response-headers", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestResponseHeaderHandler_Delete_Success(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	_ = reg.Set("http://backend:9000", upstream.ResponseHeaderRules{
		Set: map[string]string{"X-Custom": "val"},
	})
	h := NewResponseHeaderHandler(reg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/response-headers?backend=http://backend:9000", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("http://backend:9000"); ok {
		t.Error("expected rules to be deleted")
	}
}

func TestResponseHeaderHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestResponseHeaderRegistry()
	h := NewResponseHeaderHandler(reg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/response-headers", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
