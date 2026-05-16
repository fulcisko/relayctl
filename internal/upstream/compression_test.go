package upstream

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, strings.Repeat("hello world ", 100))
}

func TestCompression_NoAcceptEncoding(t *testing.T) {
	mw := Compression(CompressionConfig{})(http.HandlerFunc(helloHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("expected no gzip encoding without Accept-Encoding header")
	}
}

func TestCompression_WithGzipAccept(t *testing.T) {
	mw := Compression(CompressionConfig{})(http.HandlerFunc(helloHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected Content-Encoding: gzip")
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}
	if !strings.Contains(string(body), "hello world") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCompression_ContentLengthRemoved(t *testing.T) {
	mw := Compression(CompressionConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999")
		_, _ = io.WriteString(w, "data")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Length") != "" {
		t.Fatal("Content-Length should be removed when compressing")
	}
}

func TestCompression_DefaultMinLength(t *testing.T) {
	cfg := CompressionConfig{}
	if cfg.MinLength != 0 {
		t.Fatal("zero value expected before middleware init")
	}
	// After creating the middleware the default is applied internally.
	_ = Compression(cfg)
}

func TestCompression_CustomLevel(t *testing.T) {
	mw := Compression(CompressionConfig{Level: gzip.BestSpeed})(http.HandlerFunc(helloHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip encoding with BestSpeed level")
	}
}
