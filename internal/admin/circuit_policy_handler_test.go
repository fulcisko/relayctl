package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lukethinker/relayctl/internal/upstream"
)

func newTestCircuitPolicyRegistry() *upstream.CircuitPolicyRegistry {
	return upstream.NewCircuitPolicyRegistry()
}

func TestCircuitPolicyHandler_Get_Empty(t *testing.T) {
	reg := newTestCircuitPolicyRegistry()
	h := NewCircuitPolicyHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/admin/circuit-policy", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rw.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestCircuitPolicyHandler_Put_Valid(t *testing.T) {
	reg := newTestCircuitPolicyRegistry()
	h := NewCircuitPolicyHandler(reg)

	payload := map[string]interface{}{
		"backend":   "http://backend:9000",
		"threshold": 3,
		"timeout":   "20s",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/admin/circuit-policy", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	p, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected policy to be stored")
	}
	if p.Threshold != 3 {
		t.Errorf("expected threshold 3, got %d", p.Threshold)
	}
	if p.Timeout != 20*time.Second {
		t.Errorf("expected timeout 20s, got %v", p.Timeout)
	}
}

func TestCircuitPolicyHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestCircuitPolicyRegistry()
	h := NewCircuitPolicyHandler(reg)

	req := httptest.NewRequest(http.MethodPut, "/admin/circuit-policy", bytes.NewBufferString("not-json"))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestCircuitPolicyHandler_Delete_Success(t *testing.T) {
	reg := newTestCircuitPolicyRegistry()
	_ = reg.Set("http://backend:9000", upstream.CircuitPolicy{Threshold: 2, Timeout: 10 * time.Second})
	h := NewCircuitPolicyHandler(reg)

	payload := map[string]string{"backend": "http://backend:9000"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodDelete, "/admin/circuit-policy", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	_, ok := reg.Get("http://backend:9000")
	if ok {
		t.Error("expected policy to be deleted")
	}
}

func TestCircuitPolicyHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestCircuitPolicyRegistry()
	h := NewCircuitPolicyHandler(reg)

	req := httptest.NewRequest(http.MethodPost, "/admin/circuit-policy", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
