package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/upstream"
)

// tagFilterHandlerState holds the mutable required-tags list served by
// NewTagFilterHandler.
type tagFilterHandlerState struct {
	balancer *upstream.TagFilterBalancer
}

// NewTagFilterHandler returns an http.Handler that exposes GET / PUT for the
// required-tags list of a TagFilterBalancer.
//
//	GET  /admin/tag-filter   → {"required":["eu","premium"]}
//	PUT  /admin/tag-filter   ← {"required":["eu"]}
func NewTagFilterHandler(b *upstream.TagFilterBalancer) http.Handler {
	s := &tagFilterHandlerState{balancer: b}
	return http.HandlerFunc(s.handle)
}

func (s *tagFilterHandlerState) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPut:
		s.handlePut(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *tagFilterHandlerState) handleGet(w http.ResponseWriter, _ *http.Request) {
	if s.balancer == nil {
		http.Error(w, "no balancer configured", http.StatusServiceUnavailable)
		return
	}
	resp := map[string][]string{"required": s.balancer.RequiredTags()}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *tagFilterHandlerState) handlePut(w http.ResponseWriter, r *http.Request) {
	if s.balancer == nil {
		http.Error(w, "no balancer configured", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Required []string `json:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	s.balancer.SetRequiredTags(body.Required)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{"required": body.Required})
}
