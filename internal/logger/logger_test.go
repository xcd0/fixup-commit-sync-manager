package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	logger, err := NewLogger("INFO", logPath, true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger.level != INFO {
		t.Errorf("Expected log level INFO, got %v", logger.level)
	}

	if !logger.verbose {
		t.Error("Expected verbose to be true")
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should be created")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"debug", DEBUG},
		{"INFO", INFO},
		{"info", INFO},
		{"WARN", WARN},
		{"warn", WARN},
		{"ERROR", ERROR},
		{"error", ERROR},
		{"invalid", INFO}, // default
		{"", INFO},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	logger, err := NewLogger("WARN", logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if strings.Contains(logContent, "debug message") {
		t.Error("Debug message should not be logged when level is WARN")
	}

	if strings.Contains(logContent, "info message") {
		t.Error("Info message should not be logged when level is WARN")
	}

	if !strings.Contains(logContent, "warn message") {
		t.Error("Warn message should be logged")
	}

	if !strings.Contains(logContent, "error message") {
		t.Error("Error message should be logged")
	}
}

func TestLoggerOperationMethods(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	logger, err := NewLogger("DEBUG", logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.LogOperationStart("test-operation")
	logger.LogOperationEnd("test-operation", time.Millisecond*100)
	logger.LogConfigLoad("/path/to/config.hjson")
	logger.LogSyncResult(2, 3, 1, "abcdef1234567890")
	logger.LogFixupResult(5, "1234567890abcdef")
	logger.LogVHDXOperation("mount", "/path/to/test.vhdx")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	expectedStrings := []string{
		"Starting operation: test-operation",
		"Completed operation: test-operation",
		"Configuration loaded from: /path/to/config.hjson",
		"Sync completed - Files: +2 ~3 -1",
		"Fixup completed - 5 files modified",
		"VHDX operation: mount",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(logContent, expected) {
			t.Errorf("Log should contain %q", expected)
		}
	}
}

func TestLoggerClose(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	logger, err := NewLogger("INFO", logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message")

	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should still exist after closing logger")
	}
}
