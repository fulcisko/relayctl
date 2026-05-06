package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	l := New(3, time.Second)
	for i := 0; i < 3; i++ {
		if !l.Allow("127.0.0.1") {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := New(2, time.Second)
	l.Allow("10.0.0.1")
	l.Allow("10.0.0.1")
	if l.Allow("10.0.0.1") {
		t.Fatal("expected deny after limit exceeded")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	l := New(1, 50*time.Millisecond)
	l.Allow("192.168.1.1")
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("192.168.1.1") {
		t.Fatal("expected allow after window reset")
	}
}

func TestAllow_DifferentKeys(t *testing.T) {
	l := New(1, time.Second)
	if !l.Allow("a") {
		t.Fatal("expected allow for key a")
	}
	if !l.Allow("b") {
		t.Fatal("expected allow for key b")
	}
}

func TestMiddleware_Allows(t *testing.T) {
	l := New(5, time.Second)
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
}

func TestMiddleware_Blocks(t *testing.T) {
	l := New(1, time.Second)
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.5.5.5:80"

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	rw2 := httptest.NewRecorder()
	handler.ServeHTTP(rw2, req)

	if rw2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rw2.Code)
	}
}

func TestMiddleware_XForwardedFor(t *testing.T) {
	l := New(1, time.Second)
	handler := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "9.9.9.9")

	rw1 := httptest.NewRecorder()
	handler.ServeHTTP(rw1, req)

	rw2 := httptest.NewRecorder()
	handler.ServeHTTP(rw2, req)

	if rw2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for forwarded IP, got %d", rw2.Code)
	}
}
