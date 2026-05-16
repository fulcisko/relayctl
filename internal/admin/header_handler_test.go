package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucianoayres/relayctl/internal/upstream"
)

func newTestHeaderRegistry() *upstream.HeaderRegistry {
	reg := upstream.NewHeaderRegistry()
	reg.Set("http://backend:9000", upstream.HeaderRules{
		RequestAdd:  map[string]string{"X-Relay": "on"},
		ResponseDel: []string{"X-Secret"},
	})
	return reg
}

func TestHeaderHandler_Get_Empty(t *testing.T) {
	reg := upstream.NewHeaderRegistry()
	h := NewHeaderHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/admin/headers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestHeaderHandler_Put_Valid(t *testing.T) {
	reg := upstream.NewHeaderRegistry()
	h := NewHeaderHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"backend": "http://backend:9000",
		"rules": map[string]interface{}{
			"request_add": map[string]string{"X-Relay": "on"},
			"response_del": []string{"X-Secret"},
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/headers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	rules, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected rules to be stored")
	}
	if rules.RequestAdd["X-Relay"] != "on" {
		t.Errorf("unexpected request_add value: %v", rules.RequestAdd)
	}
}

func TestHeaderHandler_Put_InvalidJSON(t *testing.T) {
	reg := upstream.NewHeaderRegistry()
	h := NewHeaderHandler(reg)

	req := httptest.NewRequest(http.MethodPut, "/admin/headers", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHeaderHandler_Put_MissingBackend(t *testing.T) {
	reg := upstream.NewHeaderRegistry()
	h := NewHeaderHandler(reg)

	body, _ := json.Marshal(map[string]interface{}{
		"backend": "",
		"rules":   map[string]interface{}{},
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/headers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHeaderHandler_Delete_Success(t *testing.T) {
	reg := newTestHeaderRegistry()
	h := NewHeaderHandler(reg)

	body, _ := json.Marshal(map[string]string{"backend": "http://backend:9000"})
	req := httptest.NewRequest(http.MethodDelete, "/admin/headers", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if _, ok := reg.Get("http://backend:9000"); ok {
		t.Error("expected backend to be removed")
	}
}

func TestHeaderHandler_MethodNotAllowed(t *testing.T) {
	reg := upstream.NewHeaderRegistry()
	h := NewHeaderHandler(reg)

	req := httptest.NewRequest(http.MethodPost, "/admin/headers", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
