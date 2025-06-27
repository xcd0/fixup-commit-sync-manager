package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestGetDevCurrentBranch(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping dev current branch test")
	}

	tempDir := t.TempDir()
	devRepo := filepath.Join(tempDir, "dev")
	
	// Dev リポジトリを作成。
	if err := createTestRepositoryDynamic(devRepo); err != nil {
		t.Fatalf("Failed to create dev repository: %v", err)
	}

	// テスト用ブランチを作成。
	cmd := exec.Command("git", "checkout", "-b", "feature-test")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test branch: %v", err)
	}

	cfg := &config.Config{
		DevRepoPath:   devRepo,
		GitExecutable: "git",
	}

	syncer := NewFileSyncer(cfg)
	
	branch, err := syncer.getDevCurrentBranch()
	if err != nil {
		t.Fatalf("getDevCurrentBranch() failed: %v", err)
	}

	if branch != "feature-test" {
		t.Errorf("Expected branch 'feature-test', got '%s'", branch)
	}
}

func TestEnsureOpsBranch(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping ops branch test")
	}

	tempDir := t.TempDir()
	opsRepo := filepath.Join(tempDir, "ops")
	
	// Ops リポジトリを作成。
	if err := createTestRepositoryDynamic(opsRepo); err != nil {
		t.Fatalf("Failed to create ops repository: %v", err)
	}

	cfg := &config.Config{
		OpsRepoPath:   opsRepo,
		GitExecutable: "git",
	}

	syncer := NewFileSyncer(cfg)
	
	// 新しいブランチに切り替えテスト。
	err := syncer.ensureOpsBranch("feature-new")
	if err != nil {
		t.Fatalf("ensureOpsBranch() failed: %v", err)
	}

	// 現在のブランチを確認。
	currentBranch, err := syncer.getOpsCurrentBranch()
	if err != nil {
		t.Fatalf("getOpsCurrentBranch() failed: %v", err)
	}

	if currentBranch != "feature-new" {
		t.Errorf("Expected branch 'feature-new', got '%s'", currentBranch)
	}
}

func TestDynamicBranchSyncFlow(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping dynamic branch sync test")
	}

	tempDir := t.TempDir()
	devRepo := filepath.Join(tempDir, "dev")
	opsRepo := filepath.Join(tempDir, "ops")
	
	// リポジトリを作成。
	if err := createTestRepositoryDynamic(devRepo); err != nil {
		t.Fatalf("Failed to create dev repository: %v", err)
	}
	if err := createTestRepositoryDynamic(opsRepo); err != nil {
		t.Fatalf("Failed to create ops repository: %v", err)
	}

	// Dev側にfeatureブランチを作成。
	cmd := exec.Command("git", "checkout", "-b", "feature-dynamic")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create feature branch in dev: %v", err)
	}

	// Dev側に最初のファイルをコミット。
	firstFile := filepath.Join(devRepo, "first.cpp")
	err := os.WriteFile(firstFile, []byte("// First file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create first file: %v", err)
	}

	cmd = exec.Command("git", "add", "first.cpp")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add first file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add first file")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit first file: %v", err)
	}

	// Dev側にテストファイルを作成してコミット（差分検出用）。
	testFile := filepath.Join(devRepo, "test.cpp")
	err = os.WriteFile(testFile, []byte("// Test file for dynamic branch sync"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.cpp")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add test file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add test file")
	cmd.Dir = devRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit test file: %v", err)
	}

	cfg := &config.Config{
		DevRepoPath:       devRepo,
		OpsRepoPath:       opsRepo,
		IncludeExtensions: []string{".cpp"},
		GitExecutable:     "git",
		CommitTemplate:    "Auto-sync test",
		AuthorName:        "Test User",
		AuthorEmail:       "test@example.com",
		PauseLockFile:     ".sync-paused", // ロックファイル名を設定。
	}

	syncer := NewFileSyncer(cfg)
	
	// 動的ブランチ追従付きの同期を実行。
	result, err := syncer.Sync()
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}

	if len(result.FilesAdded) == 0 {
		t.Error("Expected files to be added")
	}

	// Ops側のブランチを確認。
	currentBranch, err := syncer.getOpsCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get ops current branch: %v", err)
	}

	if currentBranch != "feature-dynamic" {
		t.Errorf("Expected ops branch 'feature-dynamic', got '%s'", currentBranch)
	}

	// Ops側にファイルが同期されたか確認。
	syncedFile := filepath.Join(opsRepo, "test.cpp")
	if _, err := os.Stat(syncedFile); os.IsNotExist(err) {
		t.Error("Test file was not synced to ops repository")
	}
}

func createTestRepositoryDynamic(repoPath string) error {
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