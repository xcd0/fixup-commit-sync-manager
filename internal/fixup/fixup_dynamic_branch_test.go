package fixup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestGetDevCurrentBranchFixup(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping dev current branch test")
	}

	tempDir := t.TempDir()
	devRepo := filepath.Join(tempDir, "dev")
	
	// Dev リポジトリを作成。
	if err := createTestRepositoryFixup(devRepo); err != nil {
		t.Fatalf("Failed to create dev repository: %v", err)
	}

	// テスト用ブランチを作成。
	cmd := exec.Command("git", "checkout", "-b", "feature-fixup")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test branch: %v", err)
	}

	cfg := &config.Config{
		DevRepoPath:   devRepo,
		GitExecutable: "git",
	}

	manager := NewFixupManager(cfg)
	
	branch, err := manager.getDevCurrentBranch()
	if err != nil {
		t.Fatalf("getDevCurrentBranch() failed: %v", err)
	}

	if branch != "feature-fixup" {
		t.Errorf("Expected branch 'feature-fixup', got '%s'", branch)
	}
}

func TestEnsureOpsBranchFixup(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping ops branch test")
	}

	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")
	
	// Ops リポジトリを作成。
	if err := createTestRepositoryFixup(opsRepo); err != nil {
		t.Fatalf("Failed to create ops repository: %v", err)
	}

	cfg := &config.Config{
		OpsRepoPath:   opsRepo,
		GitExecutable: "git",
	}

	manager := NewFixupManager(cfg)
	
	// 新しいブランチに切り替えテスト。
	err := manager.ensureOpsBranch("feature-fixup-new")
	if err != nil {
		t.Fatalf("ensureOpsBranch() failed: %v", err)
	}

	// 現在のブランチを確認。
	currentBranch, err := manager.getCurrentBranch()
	if err != nil {
		t.Fatalf("getCurrentBranch() failed: %v", err)
	}

	if currentBranch != "feature-fixup-new" {
		t.Errorf("Expected branch 'feature-fixup-new', got '%s'", currentBranch)
	}
}

func TestDynamicBranchFixupFlow(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping dynamic branch fixup test")
	}

	tempDir := t.TempDir()
	devRepo := filepath.Join(tempDir, "dev")
	opsRepo := filepath.Join(tempDir, "ops")
	
	// リポジトリを作成。
	if err := createTestRepositoryFixup(devRepo); err != nil {
		t.Fatalf("Failed to create dev repository: %v", err)
	}
	if err := createTestRepositoryFixup(opsRepo); err != nil {
		t.Fatalf("Failed to create ops repository: %v", err)
	}

	// Dev側にfeatureブランチを作成。
	cmd := exec.Command("git", "checkout", "-b", "feature-fixup-flow")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create feature branch in dev: %v", err)
	}

	// Ops側にも同じブランチを作成して変更を加える。
	cmd = exec.Command("git", "checkout", "-b", "feature-fixup-flow")
	cmd.Dir = opsRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create feature branch in ops: %v", err)
	}

	// 最初のコミットを作成。
	testFile := filepath.Join(opsRepo, "fixup-test.cpp")
	err := os.WriteFile(testFile, []byte("// Initial content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "fixup-test.cpp")
	cmd.Dir = opsRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial fixup test")
	cmd.Dir = opsRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit test file: %v", err)
	}

	// ファイルを変更（未コミット状態）。
	err = os.WriteFile(testFile, []byte("// Modified content for fixup"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	cfg := &config.Config{
		DevRepoPath:       devRepo,
		OpsRepoPath:       opsRepo,
		GitExecutable:     "git",
		FixupMsgPrefix:    "fixup! ",
		AutosquashEnabled: false, // テスト環境ではautosquashを無効化。
		AuthorName:        "Test User",
		AuthorEmail:       "test@example.com",
	}

	manager := NewFixupManager(cfg)
	
	// 動的ブランチ追従付きのfixupを実行。
	result, err := manager.RunFixup()
	if err != nil {
		t.Fatalf("RunFixup() failed: %v", err)
	}

	if result.FilesModified == 0 {
		t.Error("Expected files to be modified")
	}

	// Ops側のブランチを確認。
	currentBranch, err := manager.getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get ops current branch: %v", err)
	}

	if currentBranch != "feature-fixup-flow" {
		t.Errorf("Expected ops branch 'feature-fixup-flow', got '%s'", currentBranch)
	}

	// fixupコミットが作成されたか確認。
	if result.FixupCommitHash == "" {
		t.Error("Expected fixup commit hash to be set")
	}
}

func createTestRepositoryFixup(repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "commit", "--allow-empty", "-m", "Initial commit"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}