package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lukeberry99/relayctl/internal/upstream"
)

func newTestWarmupRegistry() *upstream.WarmupRegistry {
	return upstream.NewWarmupRegistry()
}

func TestWarmupHandler_Get_Empty(t *testing.T) {
	reg := newTestWarmupRegistry()
	h := NewWarmupHandler(reg)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/warmup", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := body["backends"]; !ok {
		t.Error("expected 'backends' key in response")
	}
}

func TestWarmupHandler_Put_Valid(t *testing.T) {
	reg := newTestWarmupRegistry()
	h := NewWarmupHandler(reg)

	body := map[string]interface{}{
		"backend":          "http://10.0.0.2:9000",
		"duration_seconds": 30,
		"start_weight":     0.05,
	}
	b, _ := json.Marshal(body)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/warmup", bytes.NewReader(b))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	w := reg.Weight("http://10.0.0.2:9000", time.Now())
	if w <= 0 || w > 1 {
		t.Errorf("unexpected weight %f", w)
	}
}

func TestWarmupHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestWarmupRegistry()
	h := NewWarmupHandler(reg)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/warmup", bytes.NewBufferString("{bad"))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestWarmupHandler_Put_MissingBackend(t *testing.T) {
	reg := newTestWarmupRegistry()
	h := NewWarmupHandler(reg)

	body := map[string]interface{}{"duration_seconds": 30}
	b, _ := json.Marshal(body)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/warmup", bytes.NewReader(b))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestWarmupHandler_Delete_Success(t *testing.T) {
	reg := newTestWarmupRegistry()
	_ = reg.Register("http://10.0.0.3:8080", 60, 0.1)
	h := NewWarmupHandler(reg)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/warmup?backend=http://10.0.0.3:8080", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestWarmupHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestWarmupRegistry()
	h := NewWarmupHandler(reg)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/warmup", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
