package fixup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestNewFixupManager(t *testing.T) {
	cfg := &config.Config{
		OpsRepoPath:       "/path/to/ops",
		TargetBranch:      "sync-branch",
		BaseBranch:        "main",
		FixupMsgPrefix:    "fixup! ",
		AutosquashEnabled: true,
	}

	manager := NewFixupManager(cfg)

	if manager.cfg != cfg {
		t.Error("FixupManager should store the provided config")
	}
}

func TestGenerateFixupMessage(t *testing.T) {
	cfg := &config.Config{
		FixupMsgPrefix: "fixup! ",
	}

	manager := NewFixupManager(cfg)
	baseCommit := "abcdef1234567890"

	message := manager.generateFixupMessage(baseCommit)

	expectedParts := []string{
		"fixup! ",
		"Automated fixup",
		"abcdef12",
	}

	for _, part := range expectedParts {
		if !contains(message, part) {
			t.Errorf("Fixup message should contain %q, got: %s", part, message)
		}
	}
}

func TestValidateRepository(t *testing.T) {
	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")
	gitDir := filepath.Join(opsRepo, ".git")

	os.MkdirAll(gitDir, 0755)

	cfg := &config.Config{
		OpsRepoPath:   opsRepo,
		TargetBranch:  "master",
		GitExecutable: "git",
	}

	manager := NewFixupManager(cfg)

	if !isGitAvailable() {
		t.Skip("Git not available, skipping repository validation test")
	}

	if err := createTestRepository(opsRepo); err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(opsRepo)

	cmd := exec.Command("git", "branch", "-M", "master")
	cmd.Run()

	err := manager.validateRepository()
	if err != nil {
		t.Errorf("validateRepository() failed: %v", err)
	}
}

func TestValidateRepositoryMissingGit(t *testing.T) {
	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")
	os.MkdirAll(opsRepo, 0755)

	cfg := &config.Config{
		OpsRepoPath: opsRepo,
	}

	manager := NewFixupManager(cfg)

	err := manager.validateRepository()
	if err == nil {
		t.Error("validateRepository() should fail when .git directory is missing")
	}
}

func TestGetModifiedFilesCount(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping test")
	}

	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")

	if err := createTestRepository(opsRepo); err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	testFile := filepath.Join(opsRepo, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	cfg := &config.Config{
		OpsRepoPath:   opsRepo,
		GitExecutable: "git",
	}

	manager := NewFixupManager(cfg)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(opsRepo)

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = opsRepo
	cmd.Run()

	count, err := manager.getModifiedFilesCount()
	if err != nil {
		t.Errorf("getModifiedFilesCount() failed: %v", err)
	}

	if count == 0 {
		t.Error("Should detect staged files")
	}
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func createTestRepository(repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(repoPath)

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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
