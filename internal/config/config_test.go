package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create test directory
	if err := os.MkdirAll(".ignoregrets", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(".ignoregrets")

	// Test loading default config when file doesn't exist
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default values
	defaultCfg := DefaultConfig()
	if !reflect.DeepEqual(cfg, defaultCfg) {
		t.Errorf("Expected default config %+v, got %+v", defaultCfg, cfg)
	}

	// Test loading custom config
	customCfg := &Config{
		Retention:    5,
		SnapshotOn:   []string{"commit", "checkout"},
		RestoreOn:    []string{"checkout"},
		HooksEnabled: true,
		Exclude:      []string{"*.log", "*.tmp"},
		Include:      []string{".env", "config.local"},
	}

	if err := SaveConfig(customCfg); err != nil {
		t.Fatalf("Failed to save custom config: %v", err)
	}

	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load custom config: %v", err)
	}

	if !reflect.DeepEqual(loadedCfg, customCfg) {
		t.Errorf("Expected custom config %+v, got %+v", customCfg, loadedCfg)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Retention:    10,
				SnapshotOn:   []string{"commit"},
				RestoreOn:    []string{"checkout"},
				HooksEnabled: false,
			},
			wantErr: false,
		},
		{
			name: "invalid retention",
			cfg: &Config{
				Retention:    0,
				SnapshotOn:   []string{"commit"},
				RestoreOn:    []string{"checkout"},
				HooksEnabled: false,
			},
			wantErr: true,
		},
		{
			name: "invalid snapshot event",
			cfg: &Config{
				Retention:    10,
				SnapshotOn:   []string{"invalid"},
				RestoreOn:    []string{"checkout"},
				HooksEnabled: false,
			},
			wantErr: true,
		},
		{
			name: "invalid restore event",
			cfg: &Config{
				Retention:    10,
				SnapshotOn:   []string{"commit"},
				RestoreOn:    []string{"invalid"},
				HooksEnabled: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	// Create test directory
	if err := os.MkdirAll(".ignoregrets", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(".ignoregrets")

	// Test saving and loading config
	cfg := &Config{
		Retention:    5,
		SnapshotOn:   []string{"commit"},
		RestoreOn:    []string{"checkout"},
		HooksEnabled: true,
		Exclude:      []string{"*.log"},
		Include:      []string{".env"},
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(".ignoregrets", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify config
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if !reflect.DeepEqual(loadedCfg, cfg) {
		t.Errorf("Expected config %+v, got %+v", cfg, loadedCfg)
	}
}
