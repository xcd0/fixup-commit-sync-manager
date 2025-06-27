package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.IncludeExtensions) == 0 {
		t.Error("Default config should have include extensions")
	}

	expectedExtensions := []string{".cpp", ".h", ".hpp"}
	for i, ext := range expectedExtensions {
		if i >= len(cfg.IncludeExtensions) || cfg.IncludeExtensions[i] != ext {
			t.Errorf("Expected extension %s at index %d", ext, i)
		}
	}

	if cfg.SyncInterval != "5m" {
		t.Errorf("Expected sync interval 5m, got %s", cfg.SyncInterval)
	}

	if cfg.FixupInterval != "1h" {
		t.Errorf("Expected fixup interval 1h, got %s", cfg.FixupInterval)
	}

	if !cfg.AutosquashEnabled {
		t.Error("Expected autosquash to be enabled by default")
	}
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.hjson")

	configContent := `{
  "devRepoPath": "/path/to/dev",
  "opsRepoPath": "/path/to/ops",
  "syncInterval": "10m",
  "fixupInterval": "2h",
  "targetBranch": "test-branch",
  "includeExtensions": [".cpp", ".h"],
  "verbose": true
}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DevRepoPath != "/path/to/dev" {
		t.Errorf("Expected dev repo path '/path/to/dev', got '%s'", cfg.DevRepoPath)
	}

	if cfg.OpsRepoPath != "/path/to/ops" {
		t.Errorf("Expected ops repo path '/path/to/ops', got '%s'", cfg.OpsRepoPath)
	}

	if cfg.SyncInterval != "10m" {
		t.Errorf("Expected sync interval '10m', got '%s'", cfg.SyncInterval)
	}

	if cfg.TargetBranch != "test-branch" {
		t.Errorf("Expected target branch 'test-branch', got '%s'", cfg.TargetBranch)
	}

	if !cfg.Verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.hjson")
	if err == nil {
		t.Error("Expected error when loading non-existent config file")
	}
}

func TestLoadConfigInvalidHJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.hjson")

	invalidContent := `{
  "devRepoPath": "/path/to/dev"
  "opsRepoPath": 
}`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid HJSON")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				DevRepoPath:  "/path/to/dev",
				OpsRepoPath:  "/path/to/ops",
				SyncInterval: "5m",
				FixupInterval: "1h",
				RetryDelay:   "30s",
				LogLevel:     "INFO",
			},
			wantErr: false,
		},
		{
			name: "missing dev repo path",
			cfg: &Config{
				OpsRepoPath:  "/path/to/ops",
				SyncInterval: "5m",
				FixupInterval: "1h",
				RetryDelay:   "30s",
				LogLevel:     "INFO",
			},
			wantErr: true,
		},
		{
			name: "missing ops repo path",
			cfg: &Config{
				DevRepoPath:  "/path/to/dev",
				SyncInterval: "5m",
				FixupInterval: "1h",
				RetryDelay:   "30s",
				LogLevel:     "INFO",
			},
			wantErr: true,
		},
		{
			name: "invalid sync interval",
			cfg: &Config{
				DevRepoPath:  "/path/to/dev",
				OpsRepoPath:  "/path/to/ops",
				SyncInterval: "invalid",
				FixupInterval: "1h",
				RetryDelay:   "30s",
				LogLevel:     "INFO",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				DevRepoPath:  "/path/to/dev",
				OpsRepoPath:  "/path/to/ops",
				SyncInterval: "5m",
				FixupInterval: "1h",
				RetryDelay:   "30s",
				LogLevel:     "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigDurationMethods(t *testing.T) {
	cfg := &Config{
		SyncInterval:  "5m",
		FixupInterval: "1h",
		RetryDelay:    "30s",
	}

	syncDuration, err := cfg.GetSyncIntervalDuration()
	if err != nil {
		t.Errorf("GetSyncIntervalDuration() failed: %v", err)
	}
	if syncDuration != 5*time.Minute {
		t.Errorf("Expected 5 minutes, got %v", syncDuration)
	}

	fixupDuration, err := cfg.GetFixupIntervalDuration()
	if err != nil {
		t.Errorf("GetFixupIntervalDuration() failed: %v", err)
	}
	if fixupDuration != time.Hour {
		t.Errorf("Expected 1 hour, got %v", fixupDuration)
	}

	retryDuration, err := cfg.GetRetryDelayDuration()
	if err != nil {
		t.Errorf("GetRetryDelayDuration() failed: %v", err)
	}
	if retryDuration != 30*time.Second {
		t.Errorf("Expected 30 seconds, got %v", retryDuration)
	}
}