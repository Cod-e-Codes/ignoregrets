package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure for ignoregrets
type Config struct {
	Retention    int      `yaml:"retention"`
	SnapshotOn   []string `yaml:"snapshot_on"`
	RestoreOn    []string `yaml:"restore_on"`
	HooksEnabled bool     `yaml:"hooks_enabled"`
	Exclude      []string `yaml:"exclude"`
	Include      []string `yaml:"include"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Retention:    10,
		SnapshotOn:   []string{"commit"},
		RestoreOn:    []string{"checkout"},
		HooksEnabled: false,
		Exclude:      []string{},
		Include:      []string{},
	}
}

// LoadConfig loads the configuration from .ignoregrets/config.yaml
func LoadConfig() (*Config, error) {
	configPath := filepath.Join(".ignoregrets", "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		cfg := DefaultConfig()
		if err := SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for unset values
	if cfg.Retention <= 0 {
		cfg.Retention = DefaultConfig().Retention
	}
	if len(cfg.SnapshotOn) == 0 {
		cfg.SnapshotOn = DefaultConfig().SnapshotOn
	}
	if len(cfg.RestoreOn) == 0 {
		cfg.RestoreOn = DefaultConfig().RestoreOn
	}

	return cfg, nil
}

// SaveConfig saves the configuration to .ignoregrets/config.yaml
func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(".ignoregrets", "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig checks if the configuration is valid
func ValidateConfig(cfg *Config) error {
	if cfg.Retention < 1 {
		return fmt.Errorf("retention must be greater than 0")
	}

	validEvents := map[string]bool{
		"commit":   true,
		"checkout": true,
	}

	for _, event := range cfg.SnapshotOn {
		if !validEvents[event] {
			return fmt.Errorf("invalid snapshot_on event: %s", event)
		}
	}

	for _, event := range cfg.RestoreOn {
		if !validEvents[event] {
			return fmt.Errorf("invalid restore_on event: %s", event)
		}
	}

	return nil
}
