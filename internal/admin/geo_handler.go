package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/upstream"
)

// GeoBalancerIface is the subset of upstream.GeoBalancer used by the handler.
type GeoBalancerIface interface {
	Backends() []string
	SetRegion(region string, b upstream.Balancer) error
}

// NewGeoHandler returns an HTTP handler that exposes geo-balancer state and
// allows updating region→backend mappings at runtime.
func NewGeoHandler(geo GeoBalancerIface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if geo == nil {
			http.Error(w, "geo balancer not configured", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetGeo(w, geo)
		case http.MethodPut:
			handleSetGeoRegion(w, r, geo)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

type geoSnapshot struct {
	Backends []string `json:"backends"`
}

func handleGetGeo(w http.ResponseWriter, geo GeoBalancerIface) {
	snap := geoSnapshot{Backends: geo.Backends()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap) //nolint:errcheck
}

type setRegionRequest struct {
	Region   string   `json:"region"`
	Backends []string `json:"backends"`
}

func handleSetGeoRegion(w http.ResponseWriter, r *http.Request, geo GeoBalancerIface) {
	var req setRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Region == "" {
		http.Error(w, "region is required", http.StatusBadRequest)
		return
	}
	if len(req.Backends) == 0 {
		http.Error(w, "backends must not be empty", http.StatusBadRequest)
		return
	}
	b, err := upstream.New(req.Backends)
	if err != nil {
		http.Error(w, "invalid backends: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := geo.SetRegion(req.Region, b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
