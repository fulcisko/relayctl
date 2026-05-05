package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Route defines a single proxy routing rule.
type Route struct {
	Name    string `yaml:"name"`
	Match   string `yaml:"match"`
	Backend string `yaml:"backend"`
	Strip   bool   `yaml:"strip_prefix"`
}

// Server holds listener configuration.
type Server struct {
	Addr string `yaml:"addr"`
	TLS  *TLS   `yaml:"tls,omitempty"`
}

// TLS holds certificate paths for HTTPS.
type TLS struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// Config is the top-level relayctl configuration.
type Config struct {
	Server Server  `yaml:"server"`
	Routes []Route `yaml:"routes"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return &cfg, nil
}

// Validate checks required fields and returns an error if the config is invalid.
func (c *Config) Validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server.addr must not be empty")
	}
	for i, r := range c.Routes {
		if r.Name == "" {
			return fmt.Errorf("route[%d]: name must not be empty", i)
		}
		if r.Match == "" {
			return fmt.Errorf("route[%d] %q: match must not be empty", i, r.Name)
		}
		if r.Backend == "" {
			return fmt.Errorf("route[%d] %q: backend must not be empty", i, r.Name)
		}
	}
	return nil
}
