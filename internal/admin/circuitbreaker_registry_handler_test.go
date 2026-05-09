package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucianoayres/relayctl/internal/circuitbreaker"
)

func newTestRegistry() *circuitbreaker.Registry {
	return circuitbreaker.NewRegistry(circuitbreaker.Config{MaxFailures: 3, OpenTimeout: 5})
}

func TestCBRegistryHandler_Snapshot_Empty(t *testing.T) {
	h := NewCircuitBreakerRegistryHandler(newTestRegistry())
	req := httptest.NewRequest(http.MethodGet, "/admin/circuit-breakers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty snapshot, got %v", result)
	}
}

func TestCBRegistryHandler_Snapshot_WithEntries(t *testing.T) {
	reg := newTestRegistry()
	reg.Get("http://backend1")
	reg.Get("http://backend2")
	h := NewCircuitBreakerRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodGet, "/admin/circuit-breakers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result) //nolint:errcheck
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestCBRegistryHandler_Reset_Success(t *testing.T) {
	reg := newTestRegistry()
	cb := reg.Get("http://backend1")
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	h := NewCircuitBreakerRegistryHandler(reg)
	req := httptest.NewRequest(http.MethodDelete, "/admin/circuit-breakers?backend=http://backend1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected Closed after reset, got %s", cb.State())
	}
}

func TestCBRegistryHandler_Reset_MissingBackend(t *testing.T) {
	h := NewCircuitBreakerRegistryHandler(newTestRegistry())
	req := httptest.NewRequest(http.MethodDelete, "/admin/circuit-breakers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCBRegistryHandler_MethodNotAllowed(t *testing.T) {
	h := NewCircuitBreakerRegistryHandler(newTestRegistry())
	req := httptest.NewRequest(http.MethodPost, "/admin/circuit-breakers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
