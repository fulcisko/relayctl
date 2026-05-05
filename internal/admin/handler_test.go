package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/relayctl/internal/config"
)

type mockReloader struct {
	err error
}

func (m *mockReloader) Reload() error { return m.err }

func testCfg() *config.Config {
	return &config.Config{
		Addr: ":8080",
		Rules: []config.Rule{
			{Path: "/api", Backend: "http://localhost:9000"},
		},
	}
}

func TestHandleStatus_OK(t *testing.T) {
	h := NewHandler("config.yaml", &mockReloader{}, testCfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", resp.Addr)
	}
	if len(resp.Routes) != 1 || resp.Routes[0].Path != "/api" {
		t.Errorf("unexpected routes: %+v", resp.Routes)
	}
}

func TestHandleStatus_MethodNotAllowed(t *testing.T) {
	h := NewHandler("config.yaml", &mockReloader{}, testCfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleReload_Success(t *testing.T) {
	h := NewHandler("config.yaml", &mockReloader{}, testCfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/reload", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHandleReload_Error(t *testing.T) {
	h := NewHandler("config.yaml", &mockReloader{err: errors.New("bad config")}, testCfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/reload", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleReload_MethodNotAllowed(t *testing.T) {
	h := NewHandler("config.yaml", &mockReloader{}, testCfg)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/reload", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
