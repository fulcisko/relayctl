package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/relayctl/internal/upstream"
)

func newTestBalancer(t *testing.T) *upstream.Balancer {
	t.Helper()
	b, err := upstream.New([]string{"http://backend1:8080", "http://backend2:8080"})
	if err != nil {
		t.Fatalf("failed to create balancer: %v", err)
	}
	return b
}

func TestBalancerHandler_Get(t *testing.T) {
	h := NewBalancerHandler(newTestBalancer(t))
	req := httptest.NewRequest(http.MethodGet, "/admin/balancer", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["count"].(float64) != 2 {
		t.Errorf("expected count 2, got %v", resp["count"])
	}
}

func TestBalancerHandler_Update_Valid(t *testing.T) {
	h := NewBalancerHandler(newTestBalancer(t))
	body := `{"backends":["http://new1:9090"]}`
	req := httptest.NewRequest(http.MethodPut, "/admin/balancer", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestBalancerHandler_Update_Empty(t *testing.T) {
	h := NewBalancerHandler(newTestBalancer(t))
	body := `{"backends":[]}`
	req := httptest.NewRequest(http.MethodPut, "/admin/balancer", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBalancerHandler_Update_InvalidJSON(t *testing.T) {
	h := NewBalancerHandler(newTestBalancer(t))
	req := httptest.NewRequest(http.MethodPut, "/admin/balancer", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBalancerHandler_MethodNotAllowed(t *testing.T) {
	h := NewBalancerHandler(newTestBalancer(t))
	req := httptest.NewRequest(http.MethodDelete, "/admin/balancer", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
