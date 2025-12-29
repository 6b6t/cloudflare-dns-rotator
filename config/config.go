package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Cloudflare API token with DNS edit permissions
	APIToken string `json:"api_token"`

	// The domain/zone to target (e.g., "example.com")
	Domain string `json:"domain"`

	// The CNAME record name to rotate (e.g., "app" for app.example.com)
	RecordName string `json:"record_name"`

	// List of target addresses to rotate between
	Targets []string `json:"targets"`

	// Rotation interval (e.g., "30s", "5m", "1h")
	Interval string `json:"interval"`

	// Whether the record should be proxied through Cloudflare
	Proxied bool `json:"proxied"`

	// TTL for the DNS record (1 = auto)
	TTL int `json:"ttl"`
}

// Duration returns the parsed interval duration
func (c *Config) Duration() (time.Duration, error) {
	return time.ParseDuration(c.Interval)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.RecordName == "" {
		return fmt.Errorf("record_name is required")
	}
	if len(c.Targets) < 2 {
		return fmt.Errorf("at least 2 targets are required for rotation")
	}
	if c.Interval == "" {
		return fmt.Errorf("interval is required")
	}
	if _, err := c.Duration(); err != nil {
		return fmt.Errorf("invalid interval format: %w", err)
	}
	if c.TTL < 1 {
		c.TTL = 1 // Auto TTL
	}
	return nil
}

// Load reads and parses the configuration from a JSON file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
