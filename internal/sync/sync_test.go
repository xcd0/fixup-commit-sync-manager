package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestNewFileSyncer(t *testing.T) {
	cfg := &config.Config{
		DevRepoPath: "/path/to/dev",
		OpsRepoPath: "/path/to/ops",
	}

	syncer := NewFileSyncer(cfg)

	if syncer.cfg != cfg {
		t.Error("FileSyncer should store the provided config")
	}
}

func TestShouldIncludeFile(t *testing.T) {
	cfg := &config.Config{
		IncludeExtensions: []string{".cpp", ".h", ".hpp"},
		IncludePatterns:   []string{"src/**/*.cpp"},
		ExcludePatterns:   []string{"bin/**", "obj/**"},
	}

	syncer := NewFileSyncer(cfg)

	tests := []struct {
		filePath string
		expected bool
	}{
		{"main.cpp", true},
		{"header.h", true},
		{"template.hpp", true},
		{"src/module/test.cpp", true},
		{"readme.txt", false},
		{"bin/output.exe", false},
		{"obj/temp.cpp", false},
		{"MAIN.CPP", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := syncer.shouldIncludeFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("shouldIncludeFile(%q) = %t, want %t", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	cfg := &config.Config{
		ExcludePatterns: []string{"bin/**", "obj/**", "*.tmp"},
	}

	syncer := NewFileSyncer(cfg)

	tests := []struct {
		filePath string
		expected bool
	}{
		{"main.cpp", false},
		{"bin/output.exe", true},
		{"obj/temp.o", true},
		{"temp.tmp", true},
		{"src/main.cpp", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := syncer.isExcluded(tt.filePath)
			if result != tt.expected {
				t.Errorf("isExcluded(%q) = %t, want %t", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestIsPaused(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		DevRepoPath:   tempDir,
		PauseLockFile: ".sync-paused",
	}

	syncer := NewFileSyncer(cfg)

	if syncer.isPaused() {
		t.Error("Should not be paused when lock file doesn't exist")
	}

	lockPath := filepath.Join(tempDir, ".sync-paused")
	os.WriteFile(lockPath, []byte("paused"), 0644)

	if !syncer.isPaused() {
		t.Error("Should be paused when lock file exists")
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	cfg := &config.Config{
		CommitTemplate: "Auto-sync: ${timestamp} @ ${hash}",
	}

	syncer := NewFileSyncer(cfg)

	changes := &SyncResult{
		FilesAdded:    []string{"file1.cpp", "file2.h"},
		FilesModified: []string{"file3.cpp"},
		FilesDeleted:  []string{"file4.h"},
		CommitHash:    "abcdef1234567890",
	}

	message := syncer.generateCommitMessage(changes)

	expectedParts := []string{
		"Auto-sync:",
		"abcdef12",
		"(4 files: +2 ~1 -1)",
	}

	for _, part := range expectedParts {
		if !contains(message, part) {
			t.Errorf("Commit message should contain %q, got: %s", part, message)
		}
	}
}

func TestFileExistsInOps(t *testing.T) {
	tempDir := t.TempDir()
	opsDir := filepath.Join(tempDir, "ops")
	os.MkdirAll(opsDir, 0755)

	testFile := filepath.Join(opsDir, "test.cpp")
	os.WriteFile(testFile, []byte("test content"), 0644)

	cfg := &config.Config{
		OpsRepoPath: opsDir,
	}

	syncer := NewFileSyncer(cfg)

	if !syncer.fileExistsInOps("test.cpp") {
		t.Error("Should detect existing file in ops repo")
	}

	if syncer.fileExistsInOps("nonexistent.cpp") {
		t.Error("Should not detect non-existent file in ops repo")
	}
}

func TestCopyFileToOps(t *testing.T) {
	tempDir := t.TempDir()
	devDir := filepath.Join(tempDir, "dev")
	opsDir := filepath.Join(tempDir, "ops")

	os.MkdirAll(devDir, 0755)
	os.MkdirAll(opsDir, 0755)

	srcFile := filepath.Join(devDir, "test.cpp")
	testContent := "test content"
	os.WriteFile(srcFile, []byte(testContent), 0644)

	cfg := &config.Config{
		DevRepoPath: devDir,
		OpsRepoPath: opsDir,
	}

	syncer := NewFileSyncer(cfg)

	err := syncer.copyFileToOps("test.cpp")
	if err != nil {
		t.Fatalf("copyFileToOps failed: %v", err)
	}

	dstFile := filepath.Join(opsDir, "test.cpp")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Copied file content = %q, want %q", string(content), testContent)
	}
}

func TestDeleteFileFromOps(t *testing.T) {
	tempDir := t.TempDir()
	opsDir := filepath.Join(tempDir, "ops")
	os.MkdirAll(opsDir, 0755)

	testFile := filepath.Join(opsDir, "test.cpp")
	os.WriteFile(testFile, []byte("test content"), 0644)

	cfg := &config.Config{
		OpsRepoPath: opsDir,
	}

	syncer := NewFileSyncer(cfg)

	err := syncer.deleteFileFromOps("test.cpp")
	if err != nil {
		t.Fatalf("deleteFileFromOps failed: %v", err)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
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
