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

// serveBalancer is a test helper that sends a request to the balancer handler
// and returns the recorded response.
func serveBalancer(t *testing.T, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewBalancerHandler(newTestBalancer(t))
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, "/admin/balancer", strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, "/admin/balancer", nil)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestBalancerHandler_Get(t *testing.T) {
	w := serveBalancer(t, http.MethodGet, "")

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
	w := serveBalancer(t, http.MethodPut, `{"backends":["http://new1:9090"]}`)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestBalancerHandler_Update_Empty(t *testing.T) {
	w := serveBalancer(t, http.MethodPut, `{"backends":[]}`)

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
	w := serveBalancer(t, http.MethodDelete, "")

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
