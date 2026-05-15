package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/relayctl/internal/upstream"
)

func newTestTLSRegistry() *upstream.TLSRegistry {
	return upstream.NewTLSRegistry()
}

func TestTLSHandler_Get_Empty(t *testing.T) {
	h := NewTLSHandler(newTestTLSRegistry())
	req := httptest.NewRequest(http.MethodGet, "/admin/tls", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var snap map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&snap)
	if len(snap) != 0 {
		t.Fatalf("expected empty snapshot")
	}
}

func TestTLSHandler_Put_Valid(t *testing.T) {
	reg := newTestTLSRegistry()
	h := NewTLSHandler(reg)
	body := `{"backend":"https://svc:443","insecure_skip_verify":true,"server_name":"svc"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/tls", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	cfg, ok := reg.Get("https://svc:443")
	if !ok {
		t.Fatal("expected config to be registered")
	}
	if cfg.ServerName != "svc" {
		t.Errorf("expected ServerName=svc, got %q", cfg.ServerName)
	}
}

func TestTLSHandler_Put_InvalidJSON(t *testing.T) {
	h := NewTLSHandler(newTestTLSRegistry())
	req := httptest.NewRequest(http.MethodPut, "/admin/tls", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTLSHandler_Put_EmptyBackend(t *testing.T) {
	h := NewTLSHandler(newTestTLSRegistry())
	body := `{"backend":"","server_name":"x"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/tls", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTLSHandler_Delete_Valid(t *testing.T) {
	reg := newTestTLSRegistry()
	_ = reg.Set("https://svc:443", upstream.TLSConfig{ServerName: "svc"})
	h := NewTLSHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/tls?backend=https://svc:443", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("https://svc:443"); ok {
		t.Fatal("expected config to be removed")
	}
}

func TestTLSHandler_Delete_MissingParam(t *testing.T) {
	h := NewTLSHandler(newTestTLSRegistry())
	req := httptest.NewRequest(http.MethodDelete, "/admin/tls", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTLSHandler_MethodNotAllowed(t *testing.T) {
	h := NewTLSHandler(newTestTLSRegistry())
	req := httptest.NewRequest(http.MethodPost, "/admin/tls", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
