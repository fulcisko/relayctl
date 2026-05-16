package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestCompressionStore() *CompressionStore {
	return NewCompressionStore()
}

func TestCompressionHandler_Get_Defaults(t *testing.T) {
	h := NewCompressionHandler(newTestCompressionStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var cfg CompressionConfig
	if err := json.NewDecoder(rec.Body).Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled=true by default")
	}
	if cfg.MinLength != 512 {
		t.Errorf("expected min_length=512, got %d", cfg.MinLength)
	}
}

func TestCompressionHandler_Put_Valid(t *testing.T) {
	store := newTestCompressionStore()
	h := NewCompressionHandler(store)

	body, _ := json.Marshal(CompressionConfig{Enabled: false, MinLength: 1024})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	cfg := store.Get()
	if cfg.Enabled {
		t.Error("expected enabled=false after update")
	}
	if cfg.MinLength != 1024 {
		t.Errorf("expected min_length=1024, got %d", cfg.MinLength)
	}
}

func TestCompressionHandler_Put_InvalidJSON(t *testing.T) {
	h := NewCompressionHandler(newTestCompressionStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/", bytes.NewBufferString("not-json")))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCompressionHandler_Put_ZeroMinLength_Defaults(t *testing.T) {
	store := newTestCompressionStore()
	h := NewCompressionHandler(store)

	body, _ := json.Marshal(CompressionConfig{Enabled: true, MinLength: 0})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.Get().MinLength != 512 {
		t.Errorf("expected min_length clamped to 512, got %d", store.Get().MinLength)
	}
}

func TestCompressionHandler_MethodNotAllowed(t *testing.T) {
	h := NewCompressionHandler(newTestCompressionStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
