package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/relayctl/internal/upstream"
)

func newTestTimeoutRegistry() *upstream.TimeoutRegistry {
	return upstream.NewTimeoutRegistry()
}

func TestTimeoutHandler_Get_Empty(t *testing.T) {
	reg := newTestTimeoutRegistry()
	h := NewTimeoutHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/admin/timeouts", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(result))
	}
}

func TestTimeoutHandler_Put_Valid(t *testing.T) {
	reg := newTestTimeoutRegistry()
	h := NewTimeoutHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"backend":            "http://svc:8080",
		"dial_ms":            50,
		"response_header_ms": 100,
		"idle_ms":            200,
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/timeouts", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	cfg, err := reg.Get("http://svc:8080")
	if err != nil {
		t.Fatalf("expected config to be stored: %v", err)
	}
	if cfg.Dial != 50*time.Millisecond {
		t.Fatalf("unexpected dial timeout: %v", cfg.Dial)
	}
}

func TestTimeoutHandler_Put_InvalidJSON(t *testing.T) {
	reg := newTestTimeoutRegistry()
	h := NewTimeoutHandler(reg)

	req := httptest.NewRequest(http.MethodPut, "/admin/timeouts", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTimeoutHandler_Delete_Success(t *testing.T) {
	reg := newTestTimeoutRegistry()
	_ = reg.Set("http://svc:8080", upstream.TimeoutConfig{Dial: 10 * time.Millisecond})
	h := NewTimeoutHandler(reg)

	req := httptest.NewRequest(http.MethodDelete, "/admin/timeouts?backend=http://svc:8080", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	_, err := reg.Get("http://svc:8080")
	if err == nil {
		t.Fatal("expected entry to be deleted")
	}
}

func TestTimeoutHandler_Delete_MissingBackend(t *testing.T) {
	reg := newTestTimeoutRegistry()
	h := NewTimeoutHandler(reg)

	req := httptest.NewRequest(http.MethodDelete, "/admin/timeouts", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTimeoutHandler_MethodNotAllowed(t *testing.T) {
	reg := newTestTimeoutRegistry()
	h := NewTimeoutHandler(reg)

	req := httptest.NewRequest(http.MethodPost, "/admin/timeouts", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
