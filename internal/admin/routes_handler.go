package admin

import (
	"encoding/json"
	"net/http"

	"github.com/user/relayctl/internal/config"
)

// RouteCollector is implemented by anything that can return the current routing config.
type RouteCollector interface {
	CurrentConfig() *config.Config
}

// routeEntry is the JSON shape returned for a single route.
type routeEntry struct {
	Path    string `json:"path"`
	Backend string `json:"backend"`
}

// routesResponse wraps the list of active routes.
type routesResponse struct {
	Addr   string       `json:"addr"`
	Routes []routeEntry `json:"routes"`
}

// NewRoutesHandler returns an http.Handler that exposes the current routing
// table as a JSON endpoint.
func NewRoutesHandler(rc RouteCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cfg := rc.CurrentConfig()
		if cfg == nil {
			http.Error(w, "no configuration loaded", http.StatusServiceUnavailable)
			return
		}

		entries := make([]routeEntry, 0, len(cfg.Rules))
		for _, rule := range cfg.Rules {
			entries = append(entries, routeEntry{
				Path:    rule.Path,
				Backend: rule.Backend,
			})
		}

		resp := routesResponse{
			Addr:   cfg.Addr,
			Routes: entries,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}
