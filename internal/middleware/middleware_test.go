package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/relayctl/internal/middleware"
)

func newLogger(buf *bytes.Buffer) *log.Logger {
	return log.New(buf, "", 0)
}

func TestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	handler := middleware.Logger(newLogger(&buf))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected method in log, got: %s", output)
	}
	if !strings.Contains(output, "/test") {
		t.Errorf("expected path in log, got: %s", output)
	}
	if !strings.Contains(output, "202") {
		t.Errorf("expected status code in log, got: %s", output)
	}
}

func TestRecovery_CatchesPanic(t *testing.T) {
	var buf bytes.Buffer
	handler := middleware.Recovery(newLogger(&buf))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(buf.String(), "test panic") {
		t.Errorf("expected panic message in log, got: %s", buf.String())
	}
}

func TestChain_OrderPreserved(t *testing.T) {
	var order []string

	mk := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := middleware.Chain(base, mk("first"), mk("second"), mk("third"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if len(order) != 3 || order[0] != "first" || order[1] != "second" || order[2] != "third" {
		t.Errorf("unexpected middleware order: %v", order)
	}
}
