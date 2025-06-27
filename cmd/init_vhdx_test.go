package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestVHDXOpsRepoPathGeneration はVHDXでのOpsリポジトリパス生成をテストする。
func TestVHDXOpsRepoPathGeneration(t *testing.T) {
	tests := []struct {
		name        string
		devRepoPath string
		mountPoint  string
		expectedBaseName string
	}{
		{
			name:        "Windows drive letter Q: with simple repo",
			devRepoPath: "/path/to/my-repo",
			mountPoint:  "Q:",
			expectedBaseName: "my-repo",
		},
		{
			name:        "Windows drive letter X: with complex path",
			devRepoPath: "C:/Users/dev/project-name", // Linux環境でもテスト可能な形式
			mountPoint:  "X:",
			expectedBaseName: "project-name",
		},
		{
			name:        "Complex repository name",
			devRepoPath: "/home/user/fixup-commit-sync-manager",
			mountPoint:  "Z:",
			expectedBaseName: "fixup-commit-sync-manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				DevRepoPath: tt.devRepoPath,
				MountPoint:  tt.mountPoint,
			}

			// init_vhdx.goの処理をシミュレート。
			devBaseName := filepath.Base(cfg.DevRepoPath)
			opsRepoPath, _ := filepath.Abs(filepath.Join(cfg.MountPoint, devBaseName))
			normalizedPath := filepath.ToSlash(opsRepoPath)

			// ベース名が正しく抽出されることを確認。
			if devBaseName != tt.expectedBaseName {
				t.Errorf("Expected base name %q, got %q", tt.expectedBaseName, devBaseName)
			}

			// パスにマウントポイントとベース名が含まれることを確認。
			if !strings.Contains(normalizedPath, tt.mountPoint) {
				t.Errorf("OpsRepoPath should contain mount point %q: %q", tt.mountPoint, normalizedPath)
			}
			if !strings.Contains(normalizedPath, tt.expectedBaseName) {
				t.Errorf("OpsRepoPath should contain base name %q: %q", tt.expectedBaseName, normalizedPath)
			}
		})
	}
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
