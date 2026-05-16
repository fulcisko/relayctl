package admin

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// CompressionConfig holds the live compression settings.
type CompressionConfig struct {
	Enabled   bool `json:"enabled"`
	MinLength int  `json:"min_length"`
}

// CompressionStore is a thread-safe holder for CompressionConfig.
type CompressionStore struct {
	val atomic.Value // stores CompressionConfig
}

// NewCompressionStore returns a store initialised with sensible defaults.
func NewCompressionStore() *CompressionStore {
	s := &CompressionStore{}
	s.val.Store(CompressionConfig{Enabled: true, MinLength: 512})
	return s
}

// Get returns the current config.
func (s *CompressionStore) Get() CompressionConfig {
	return s.val.Load().(CompressionConfig)
}

// Set replaces the current config.
func (s *CompressionStore) Set(c CompressionConfig) {
	if c.MinLength <= 0 {
		c.MinLength = 512
	}
	s.val.Store(c)
}

// NewCompressionHandler returns an HTTP handler that exposes GET / PUT
// for the compression configuration.
func NewCompressionHandler(store *CompressionStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(store.Get())

		case http.MethodPut:
			var cfg CompressionConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			store.Set(cfg)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(store.Get())

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
