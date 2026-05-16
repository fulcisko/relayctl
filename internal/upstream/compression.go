package upstream

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CompressionConfig holds settings for response compression.
type CompressionConfig struct {
	// Level is the gzip compression level (1–9). Defaults to gzip.DefaultCompression.
	Level int
	// MinLength is the minimum response body size in bytes before compression
	// is applied. Defaults to 1024.
	MinLength int
}

type compressionWriter struct {
	http.ResponseWriter
	gz     *gzip.Writer
	written int
	status  int
}

func (cw *compressionWriter) WriteHeader(code int) {
	cw.status = code
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *compressionWriter) Write(b []byte) (int, error) {
	cw.written += len(b)
	return cw.gz.Write(b)
}

var gzipPool = sync.Pool{
	New: func() any {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return gz
	},
}

// Compression returns middleware that gzip-compresses responses when the
// client sends an Accept-Encoding: gzip header.
func Compression(cfg CompressionConfig) func(http.Handler) http.Handler {
	if cfg.MinLength <= 0 {
		cfg.MinLength = 1024
	}
	level := cfg.Level
	if level == 0 {
		level = gzip.DefaultCompression
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				_ = gz.Close()
				gzipPool.Put(gz)
			}()
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			cw := &compressionWriter{ResponseWriter: w, gz: gz}
			next.ServeHTTP(cw, r)
		})
	}
}
