package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukedever/relayctl/internal/upstream"
)

func newTestHedgeRegistry() *upstream.HedgeRegistry {
	return upstream.NewHedgeRegistry()
}

func TestHedgeHandler_Get_Empty(t *testing.T) {
	h := NewHedgeHandler(newTestHedgeRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/hedge", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 0 {
		t.Errorf("expected empty map, got %v", out)
	}
}

func TestHedgeHandler_Put_Valid(t *testing.T) {
	reg := newTestHedgeRegistry()
	h := NewHedgeHandler(reg)
	body, _ := json.Marshal(map[string]interface{}{
		"backend":    "http://svc:8080",
		"delay_ms":   25,
		"max_hedges": 2,
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/hedge", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	cfg, ok := reg.Get("http://svc:8080")
	if !ok {
		t.Fatal("expected entry to be stored")
	}
	if cfg.MaxHedges != 2 {
		t.Errorf("expected MaxHedges=2, got %d", cfg.MaxHedges)
	}
}

func TestHedgeHandler_Put_InvalidJSON(t *testing.T) {
	h := NewHedgeHandler(newTestHedgeRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/hedge", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHedgeHandler_Put_MissingBackend(t *testing.T) {
	h := NewHedgeHandler(newTestHedgeRegistry())
	body, _ := json.Marshal(map[string]interface{}{"delay_ms": 10})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/hedge", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHedgeHandler_Delete_Success(t *testing.T) {
	reg := newTestHedgeRegistry()
	_ = reg.Set("http://del:80", upstream.HedgeConfig{})
	h := NewHedgeHandler(reg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/hedge?backend=http://del:80", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("http://del:80"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestHedgeHandler_MethodNotAllowed(t *testing.T) {
	h := NewHedgeHandler(newTestHedgeRegistry())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/hedge", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
