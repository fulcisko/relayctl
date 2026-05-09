package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/relayctl/internal/admin"
	"github.com/user/relayctl/internal/retry"
)

func newTestRetryHandler() http.Handler {
	policy := retry.DefaultPolicy()
	return admin.NewRetryHandler(policy)
}

func TestRetryHandler_GetPolicy(t *testing.T) {
	h := newTestRetryHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/retry", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := body["max_attempts"]; !ok {
		t.Error("expected max_attempts in response")
	}
}

func TestRetryHandler_SetPolicy_Valid(t *testing.T) {
	h := newTestRetryHandler()
	body := `{"max_attempts":5,"retry_on":[500,502],"backoff_ms":200}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/retry", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRetryHandler_SetPolicy_InvalidJSON(t *testing.T) {
	h := newTestRetryHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/retry", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRetryHandler_MethodNotAllowed(t *testing.T) {
	h := newTestRetryHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/retry", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestRetryHandler_SetPolicy_ZeroAttempts(t *testing.T) {
	h := newTestRetryHandler()
	body := `{"max_attempts":0,"retry_on":[500],"backoff_ms":100}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/retry", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero attempts, got %d", rr.Code)
	}
}
