package admin

import (
	"encoding/json"
	"net/http"

	"github.com/relayctl/relayctl/internal/config"
)

// StatusResponse represents the current proxy status.
type StatusResponse struct {
	Addr   string        `json:"addr"`
	Routes []RouteStatus `json:"routes"`
}

// RouteStatus represents a single routing rule.
type RouteStatus struct {
	Path    string `json:"path"`
	Backend string `json:"backend"`
}

// Reloader is the interface for triggering a config reload.
type Reloader interface {
	Reload() error
}

// Handler holds dependencies for the admin HTTP handlers.
type Handler struct {
	cfgPath string
	reloader Reloader
	getConfig func() *config.Config
}

// NewHandler creates a new admin Handler.
func NewHandler(cfgPath string, reloader Reloader, getConfig func() *config.Config) *Handler {
	return &Handler{
		cfgPath:   cfgPath,
		reloader:  reloader,
		getConfig: getConfig,
	}
}

// RegisterRoutes registers admin endpoints on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/status", h.handleStatus)
	mux.HandleFunc("/admin/reload", h.handleReload)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.getConfig()
	resp := StatusResponse{Addr: cfg.Addr}
	for _, rule := range cfg.Rules {
		resp.Routes = append(resp.Routes, RouteStatus{
			Path:    rule.Path,
			Backend: rule.Backend,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.reloader.Reload(); err != nil {
		http.Error(w, "reload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
