package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasl-dev/relayctl/internal/upstream"
)

func newTestBurstBalancer(t *testing.T) *upstream.BurstBalancer {
	t.Helper()
	primary, err := upstream.New([]string{"http://primary:8080"})
	if err != nil {
		t.Fatalf("primary balancer: %v", err)
	}
	burst, err := upstream.New([]string{"http://burst:9090"})
	if err != nil {
		t.Fatalf("burst balancer: %v", err)
	}
	b, err := upstream.NewBurstBalancer(primary, burst, 10, 5)
	if err != nil {
		t.Fatalf("burst balancer: %v", err)
	}
	return b
}

func TestBurstHandler_Get(t *testing.T) {
	b := newTestBurstBalancer(t)
	h := NewBurstHandler(b)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/burst", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := out["threshold"]; !ok {
		t.Error("expected threshold field in response")
	}
}

func TestBurstHandler_Put_Valid(t *testing.T) {
	b := newTestBurstBalancer(t)
	h := NewBurstHandler(b)

	body, _ := json.Marshal(map[string]any{"threshold": 20, "max_burst": 8})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/burst", bytes.NewReader(body))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBurstHandler_Put_InvalidJSON(t *testing.T) {
	b := newTestBurstBalancer(t)
	h := NewBurstHandler(b)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/burst", bytes.NewBufferString("not-json"))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBurstHandler_MethodNotAllowed(t *testing.T) {
	b := newTestBurstBalancer(t)
	h := NewBurstHandler(b)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/burst", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
