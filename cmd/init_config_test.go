package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestGenerateHJSONTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DevRepoPath = "/path/to/dev"
	cfg.OpsRepoPath = "/path/to/ops"
	cfg.VHDXPath = "/path/to/test.vhdx"
	cfg.MountPoint = "X:"

	template := generateHJSONTemplate(cfg)

	if template == "" {
		t.Error("Generated template should not be empty")
	}

	expectedStrings := []string{
		"/path/to/dev",
		"/path/to/ops",
		"Auto-sync: ${timestamp} @ ${hash}",
		"fixup! ",
		"dynamic", // ブランチは動的追従
		"/path/to/test.vhdx",
		"X:",
	}

	for _, expected := range expectedStrings {
		if !contains(template, expected) {
			t.Errorf("Template should contain %q", expected)
		}
	}
}

func TestWriteConfigTemplate(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.hjson")

	cfg := config.DefaultConfig()
	cfg.DevRepoPath = "/test/dev"
	cfg.OpsRepoPath = "/test/ops"

	err := writeConfigTemplate(configPath, cfg)
	if err != nil {
		t.Fatalf("Failed to write config template: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Config file should not be empty")
	}

	if !contains(string(content), "/test/dev") {
		t.Error("Config file should contain dev repo path")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
