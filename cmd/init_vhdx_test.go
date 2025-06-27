package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestCloneRepository(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping test")
	}

	tempDir := t.TempDir()
	sourceRepo := filepath.Join(tempDir, "source")
	targetRepo := filepath.Join(tempDir, "target")

	if err := createTestRepository(sourceRepo); err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	err := cloneRepository(sourceRepo, targetRepo, "git")
	if err != nil {
		t.Errorf("cloneRepository() failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetRepo, ".git")); os.IsNotExist(err) {
		t.Error("Cloned repository should have .git directory")
	}
}

func TestSetupOpsRepository(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping test")
	}

	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")

	if err := createTestRepository(opsRepo); err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(opsRepo)

	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/example/repo.git")
	cmd.Run()

	cfg := &config.Config{
		GitExecutable: "git",
		DevRepoPath:   "/tmp/dev",
		OpsRepoPath:   opsRepo,
	}

	err := setupOpsRepository(opsRepo, cfg)
	if err != nil {
		t.Errorf("setupOpsRepository() failed: %v", err)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping test")
	}

	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "test-repo")

	if err := createTestRepository(repoDir); err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(repoDir)

	branch, err := getCurrentBranch("git")
	if err != nil {
		t.Errorf("getCurrentBranch() failed: %v", err)
	}

	if branch == "" {
		t.Error("Branch name should not be empty")
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
