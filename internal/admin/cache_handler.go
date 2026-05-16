package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/relayctl/relayctl/internal/upstream"
)

// NewCacheHandler returns an http.Handler for inspecting and managing
// the response cache. Supports GET (stats), DELETE /flush (flush all),
// and DELETE ?key=<k> (evict single entry).
func NewCacheHandler(cache *upstream.ResponseCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cache == nil {
			http.Error(w, "cache not configured", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetCache(w, cache)
		case http.MethodDelete:
			handleDeleteCache(w, r, cache)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type cacheStats struct {
	Entries int    `json:"entries"`
	TTL     string `json:"ttl"`
}

func handleGetCache(w http.ResponseWriter, cache *upstream.ResponseCache) {
	stats := cacheStats{
		Entries: cache.Len(),
		TTL:     cache.TTL().Round(time.Millisecond).String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleDeleteCache(w http.ResponseWriter, r *http.Request, cache *upstream.ResponseCache) {
	if r.URL.Path == "/flush" || r.URL.Query().Get("flush") == "true" {
		cache.Flush()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key query param", http.StatusBadRequest)
		return
	}
	cache.Delete(key)
	w.WriteHeader(http.StatusNoContent)
}
