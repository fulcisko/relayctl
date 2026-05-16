package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/internal/upstream"
)

func newTestShadowRegistry() *upstream.ShadowRegistry {
	return upstream.NewShadowRegistry()
}

func TestShadowingHandler_Get_Empty(t *testing.T) {
	h := NewShadowingHandler(newTestShadowRegistry())
	req := httptest.NewRequest(http.MethodGet, "/admin/shadowing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 0 {
		t.Errorf("expected empty map, got %v", out)
	}
}

func TestShadowingHandler_Put_Valid(t *testing.T) {
	reg := newTestShadowRegistry()
	h := NewShadowingHandler(reg)
	body := `{"backend":"http://primary:8080","shadow":{"backend":"http://shadow:9090","sample_rate":0.5,"enabled":true}}`
	req := httptest.NewRequest(http.MethodPut, "/admin/shadowing", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	entry, ok := reg.Get("http://primary:8080")
	if !ok {
		t.Fatal("expected entry to be stored")
	}
	if entry.SampleRate != 0.5 {
		t.Errorf("expected 0.5, got %v", entry.SampleRate)
	}
}

func TestShadowingHandler_Put_InvalidJSON(t *testing.T) {
	h := NewShadowingHandler(newTestShadowRegistry())
	req := httptest.NewRequest(http.MethodPut, "/admin/shadowing", bytes.NewBufferString(`{bad}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestShadowingHandler_Put_MissingBackend(t *testing.T) {
	h := NewShadowingHandler(newTestShadowRegistry())
	body := `{"backend":"","shadow":{"sample_rate":0.1}}`
	req := httptest.NewRequest(http.MethodPut, "/admin/shadowing", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestShadowingHandler_Delete_Success(t *testing.T) {
	reg := newTestShadowRegistry()
	_ = reg.Set("http://x:1", upstream.ShadowEntry{SampleRate: 0.3, Enabled: true})
	h := NewShadowingHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/shadowing?backend=http://x:1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := reg.Get("http://x:1"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestShadowingHandler_MethodNotAllowed(t *testing.T) {
	h := NewShadowingHandler(newTestShadowRegistry())
	req := httptest.NewRequest(http.MethodPost, "/admin/shadowing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
