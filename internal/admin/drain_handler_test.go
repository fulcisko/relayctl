package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/relayctl/internal/admin"
	"github.com/yourusername/relayctl/internal/upstream"
)

func newTestDrainRegistry() *upstream.DrainRegistry {
	return upstream.NewDrainRegistry()
}

func TestDrainHandler_Get_Empty(t *testing.T) {
	h := admin.NewDrainHandler(newTestDrainRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/drain", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := body["draining"]; !ok {
		t.Error("expected 'draining' key in response")
	}
}

func TestDrainHandler_Drain_Success(t *testing.T) {
	reg := newTestDrainRegistry()
	h := admin.NewDrainHandler(reg)
	body := `{"backend":"http://backend1:8080","timeout_seconds":5}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/drain", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !reg.IsDraining("http://backend1:8080") {
		t.Error("expected backend to be marked as draining")
	}
}

func TestDrainHandler_Drain_InvalidJSON(t *testing.T) {
	h := admin.NewDrainHandler(newTestDrainRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/drain", strings.NewReader(`{bad json`))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDrainHandler_Drain_MissingBackend(t *testing.T) {
	h := admin.NewDrainHandler(newTestDrainRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/drain", strings.NewReader(`{"timeout_seconds":5}`))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDrainHandler_Restore_Success(t *testing.T) {
	reg := newTestDrainRegistry()
	reg.Drain("http://backend2:9090", 10*time.Second)
	h := admin.NewDrainHandler(reg)
	body := `{"backend":"http://backend2:9090"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/drain", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if reg.IsDraining("http://backend2:9090") {
		t.Error("expected backend to no longer be draining")
	}
}

func TestDrainHandler_MethodNotAllowed(t *testing.T) {
	h := admin.NewDrainHandler(newTestDrainRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/drain", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
