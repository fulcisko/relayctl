package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/relayctl/internal/upstream"
)

func newTestRateLimitRegistry() *upstream.RateLimitRegistry {
	return upstream.NewRateLimitRegistry()
}

func TestRateLimitRegistryHandler_Get_Empty(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodGet, "/admin/ratelimit-registry", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestRateLimitRegistryHandler_Put_Valid(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	body := `{"backend":"http://svc:9000","requests_per_second":50,"burst":100}`
	req := httptest.NewRequest(http.MethodPut, "/admin/ratelimit-registry", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	cfg, ok := reg.Get("http://svc:9000")
	if !ok {
		t.Fatal("expected entry to be stored")
	}
	if cfg.RequestsPerSecond != 50 || cfg.Burst != 100 {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestRateLimitRegistryHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodPut, "/admin/ratelimit-registry", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateLimitRegistryHandler_Put_MissingBackend(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	body := `{"backend":"","requests_per_second":10,"burst":5}`
	req := httptest.NewRequest(http.MethodPut, "/admin/ratelimit-registry", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateLimitRegistryHandler_Delete_Success(t *testing.T) {
	reg := newTestRateLimitRegistry()
	_ = reg.Set("http://svc:9000", upstream.PerBackendRateLimit{RequestsPerSecond: 5, Burst: 10})
	h := NewRateLimitRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/ratelimit-registry?backend=http://svc:9000", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, ok := reg.Get("http://svc:9000"); ok {
		t.Error("expected entry to be deleted")
	}
}

func TestRateLimitRegistryHandler_Delete_MissingParam(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/ratelimit-registry", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRateLimitRegistryHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestRateLimitRegistry()
	h := NewRateLimitRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodPatch, "/admin/ratelimit-registry", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
