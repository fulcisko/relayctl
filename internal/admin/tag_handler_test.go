package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lukasgolino/relayctl/internal/upstream"
)

func newTestTagRegistry() *upstream.TagRegistry {
	return upstream.NewTagRegistry()
}

func TestTagHandler_Get_Empty(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/tags", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty map, got %v", body)
	}
}

func TestTagHandler_Put_Valid(t *testing.T) {
	reg := newTestTagRegistry()
	h := NewTagHandler(reg)
	body := `{"backend":"http://backend:9000","tags":["stable","eu"]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/tags", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	tags, ok := reg.Get("http://backend:9000")
	if !ok {
		t.Fatal("expected tags to be set")
	}
	if len(tags) != 2 || tags[0] != "stable" || tags[1] != "eu" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestTagHandler_Put_InvalidJSON(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/tags", strings.NewReader("not-json"))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTagHandler_Put_MissingBackend(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	body := `{"backend":"","tags":["v1"]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/tags", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTagHandler_Delete_Success(t *testing.T) {
	reg := newTestTagRegistry()
	_ = reg.Set("http://backend:9000", []string{"canary"})
	h := NewTagHandler(reg)
	body := `{"backend":"http://backend:9000"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/admin/tags", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := reg.Get("http://backend:9000")
	if ok {
		t.Error("expected tag entry to be deleted")
	}
}

func TestTagHandler_MethodNotAllowed(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/tags", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
