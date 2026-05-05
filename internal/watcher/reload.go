package watcher

import (
	"fmt"
	"log"

	"github.com/user/relayctl/internal/config"
)

// Reloader holds the dependencies needed to perform a hot-reload.
type Reloader struct {
	configPath string
	onReload   func(cfg *config.Config) error
}

// NewReloader creates a Reloader that reads configPath and calls onReload
// with the new configuration on each successful parse.
func NewReloader(configPath string, onReload func(cfg *config.Config) error) *Reloader {
	return &Reloader{
		configPath: configPath,
		onReload:   onReload,
	}
}

// Reload loads the config file and invokes the reload callback.
// Errors are logged but do not stop future reload attempts.
func (r *Reloader) Reload() error {
	cfg, err := config.Load(r.configPath)
	if err != nil {
		return fmt.Errorf("reload: failed to load config: %w", err)
	}
	if err := r.onReload(cfg); err != nil {
		return fmt.Errorf("reload: callback error: %w", err)
	}
	log.Printf("[reloader] config reloaded from %s", r.configPath)
	return nil
}

// MustReload calls Reload and logs any error without returning it,
// suitable for use as a watcher onChange callback.
func (r *Reloader) MustReload() {
	if err := r.Reload(); err != nil {
		log.Printf("[reloader] error: %v", err)
	}
}
