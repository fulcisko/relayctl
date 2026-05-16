package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

func newTestQuotaRegistry() *upstream.QuotaRegistry {
	return upstream.NewQuotaRegistry()
}

func TestQuotaHandler_Get_Empty(t *testing.T) {
	h := NewQuotaHandler(newTestQuotaRegistry())
	req := httptest.NewRequest(http.MethodGet, "/admin/quota", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestQuotaHandler_Put_Valid(t *testing.T) {
	reg := newTestQuotaRegistry()
	h := NewQuotaHandler(reg)
	body := `{"backend":"http://backend:9000","max_requests":500,"window":"1m"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/quota", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	cfg, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected quota entry to exist")
	}
	if cfg.MaxRequests != 500 {
		t.Errorf("expected max_requests=500, got %d", cfg.MaxRequests)
	}
	if cfg.Window != time.Minute {
		t.Errorf("expected window=1m, got %v", cfg.Window)
	}
}

func TestQuotaHandler_Put_InvalidJSON(t *testing.T) {
	h := NewQuotaHandler(newTestQuotaRegistry())
	req := httptest.NewRequest(http.MethodPut, "/admin/quota", strings.NewReader(`{bad json`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestQuotaHandler_Put_MissingBackend(t *testing.T) {
	h := NewQuotaHandler(newTestQuotaRegistry())
	body := `{"max_requests":100,"window":"30s"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/quota", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestQuotaHandler_Delete_Success(t *testing.T) {
	reg := newTestQuotaRegistry()
	_ = reg.Set("http://backend:9001", upstream.QuotaConfig{MaxRequests: 100, Window: time.Minute})
	h := NewQuotaHandler(reg)
	body := `{"backend":"http://backend:9001"}`
	req := httptest.NewRequest(http.MethodDelete, "/admin/quota", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if _, ok := reg.Get("http://backend:9001"); ok {
		t.Error("expected quota entry to be removed")
	}
}

func TestQuotaHandler_MethodNotAllowed(t *testing.T) {
	h := NewQuotaHandler(newTestQuotaRegistry())
	req := httptest.NewRequest(http.MethodPost, "/admin/quota", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
