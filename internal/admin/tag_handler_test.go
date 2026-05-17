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

// makeTagRequest is a helper that creates a recorded HTTP request against the
// TagHandler and returns the response recorder for assertion.
func makeTagRequest(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	var reqBody *strings.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	} else {
		reqBody = strings.NewReader("")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, reqBody)
	h.ServeHTTP(rec, req)
	return rec
}

func TestTagHandler_Get_Empty(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	rec := makeTagRequest(h, http.MethodGet, "/admin/tags", "")
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
	rec := makeTagRequest(h, http.MethodPut, "/admin/tags", `{"backend":"http://backend:9000","tags":["stable","eu"]}`)
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
	rec := makeTagRequest(h, http.MethodPut, "/admin/tags", "not-json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTagHandler_Put_MissingBackend(t *testing.T) {
	h := NewTagHandler(newTestTagRegistry())
	rec := makeTagRequest(h, http.MethodPut, "/admin/tags", `{"backend":"","tags":["v1"]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTagHandler_Delete_Success(t *testing.T) {
	reg := newTestTagRegistry()
	_ = reg.Set("http://backend:9000", []string{"canary"})
	h := NewTagHandler(reg)
	rec := makeTagRequest(h, http.MethodDelete, "/admin/tags", `{"backend":"http://backend:9000"}`)
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
	rec := makeTagRequest(h, http.MethodPost, "/admin/tags", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
