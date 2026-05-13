package accesslog_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/relayctl/internal/accesslog"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	l := accesslog.New(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLog_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := accesslog.New(&buf)

	entry := accesslog.Entry{
		Timestamp:  "2024-01-01T00:00:00Z",
		Method:     "GET",
		Path:       "/health",
		Status:     200,
		DurationMs: 5,
		ClientIP:   "127.0.0.1:9000",
		Backend:    "http://backend:8080",
	}

	if err := l.Log(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got accesslog.Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if got.Method != "GET" {
		t.Errorf("expected GET, got %s", got.Method)
	}
	if got.Status != 200 {
		t.Errorf("expected 200, got %d", got.Status)
	}
}

func TestMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	l := accesslog.New(&buf)

	handler := l.Middleware("http://backend:9090", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var entry accesslog.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Status != http.StatusCreated {
		t.Errorf("logged status: want 201, got %d", entry.Status)
	}
	if entry.Method != http.MethodPost {
		t.Errorf("logged method: want POST, got %s", entry.Method)
	}
	if !strings.Contains(entry.Backend, "9090") {
		t.Errorf("backend not logged correctly: %s", entry.Backend)
	}
}

func TestMiddleware_DefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	l := accesslog.New(&buf)

	handler := l.Middleware("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no explicit WriteHeader — should default to 200
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry accesslog.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Status != 200 {
		t.Errorf("expected default status 200, got %d", entry.Status)
	}
}
