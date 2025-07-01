package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/sync"
	"fixup-commit-sync-manager/internal/fixup"
)

// TestE2ECommandExecution はrunCommand実装を使ったEnd-to-Endテスト。
func TestE2ECommandExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped in short mode")
	}

	// Step 1: 基本的なコマンド実行テスト。
	testStep_BasicCommandExecution(t)
	
	// Step 2: Git操作のコマンド実行テスト。
	testStep_GitCommandExecution(t)
	
	// Step 3: ファイル操作のコマンド実行テスト。
	testStep_FileOperationCommandExecution(t)
	
	// Step 4: エラーハンドリングのテスト。
	testStep_CommandErrorHandling(t)
}

// TestE2ERealRepositoryWorkflow は実際のGitリポジトリを使った統合テスト。
func TestE2ERealRepositoryWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped in short mode")
	}
	
	if !isGitAvailable() {
		t.Skip("Git not available, skipping real repository workflow test")
	}

	// Step 1: リポジトリ作成と初期化。
	devRepo, opsRepo := testStep_CreateRealRepositories(t)
	defer testStep_CleanupRealRepositories(t, devRepo, opsRepo)
	
	// Step 2: 設定ファイル作成。
	cfg := testStep_CreateRealConfig(t, devRepo, opsRepo)
	
	// Step 3: 実際のSync機能テスト。
	testStep_RealSyncOperation(t, cfg)
	
	// Step 4: 実際のFixup機能テスト。
	testStep_RealFixupOperation(t, cfg)
	
	// Step 5: 統合ワークフローテスト。
	testStep_RealIntegratedWorkflow(t, cfg, devRepo, opsRepo)
}

// TestE2ECompleteWorkflow は完全なワークフロー統合テスト。
func TestE2ECompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped in short mode")
	}
	
	if !isGitAvailable() {
		t.Skip("Git not available, skipping complete workflow test")
	}

	// 30秒のタイムアウトコンテキスト。
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: 完全な環境セットアップ。
	devRepo, opsRepo := testStep_CompleteEnvironmentSetup(t)
	defer testStep_CompleteCleanup(t, devRepo, opsRepo)
	
	// Step 2: 設定とマネージャー初期化。
	cfg, syncMgr, fixupMgr := testStep_InitializeManagers(t, devRepo, opsRepo)
	
	// Step 3: 継続的な開発シミュレーション。
	testStep_SimulateContinuousDevelopment(t, ctx, devRepo, syncMgr, fixupMgr)
	
	// Step 4: 結果検証。
	testStep_VerifyWorkflowResults(t, cfg, devRepo, opsRepo)
}

// =============================================================================
// Step Functions for TestE2ECommandExecution
// =============================================================================

func testStep_BasicCommandExecution(t *testing.T) {
	t.Log("Testing basic command execution...")
	
	// echoコマンドのテスト。
	err := runCommandE2E("echo 'Hello Integration Test'")
	if err != nil {
		t.Fatalf("Basic echo command failed: %v", err)
	}
	
	// dateコマンドのテスト（OS別）。
	var dateCmd string
	if IsWindowsTestEnvironment() {
		dateCmd = "date /t"
	} else {
		dateCmd = "date"
	}
	
	err = runCommandE2E(dateCmd)
	if err != nil {
		t.Fatalf("Date command failed: %v", err)
	}
	
	t.Log("✓ Basic command execution successful")
}

func testStep_GitCommandExecution(t *testing.T) {
	t.Log("Testing Git command execution...")
	
	tempDir := t.TempDir()
	testRepo := filepath.Join(tempDir, "git-test-repo")
	
	// Git repository作成。
	err := runCommandInDirE2E("git init", testRepo)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("Git not available for command execution test")
		}
		t.Fatalf("Git init failed: %v", err)
	}
	
	// Git設定。
	err = runCommandInDirE2E("git config user.name TestUser", testRepo)
	if err != nil {
		t.Fatalf("Git config name failed: %v", err)
	}
	
	err = runCommandInDirE2E("git config user.email test@example.com", testRepo)
	if err != nil {
		t.Fatalf("Git config email failed: %v", err)
	}
	
	// ファイル作成とコミット。
	testFile := filepath.Join(testRepo, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	err = runCommandInDirE2E("git add test.txt", testRepo)
	if err != nil {
		t.Fatalf("Git add failed: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m TestCommit", testRepo)
	if err != nil {
		t.Fatalf("Git commit failed: %v", err)
	}
	
	t.Log("✓ Git command execution successful")
}

func testStep_FileOperationCommandExecution(t *testing.T) {
	t.Log("Testing file operation command execution...")
	
	tempDir := t.TempDir()
	
	// ファイル作成テスト。
	testFile := filepath.Join(tempDir, "cmdtest.txt")
	var createCmd string
	if IsWindowsTestEnvironment() {
		createCmd = fmt.Sprintf("echo test content > %s", testFile)
	} else {
		createCmd = fmt.Sprintf("echo 'test content' > %s", testFile)
	}
	
	err := runCommandE2E(createCmd)
	if err != nil {
		t.Fatalf("File creation command failed: %v", err)
	}
	
	// ファイル存在確認。
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("File was not created by command: %s", testFile)
	}
	
	// ファイル削除テスト。
	var deleteCmd string
	if IsWindowsTestEnvironment() {
		deleteCmd = fmt.Sprintf("del %s", testFile)
	} else {
		deleteCmd = fmt.Sprintf("rm %s", testFile)
	}
	
	err = runCommandE2E(deleteCmd)
	if err != nil {
		t.Fatalf("File deletion command failed: %v", err)
	}
	
	t.Log("✓ File operation command execution successful")
}

func testStep_CommandErrorHandling(t *testing.T) {
	t.Log("Testing command error handling...")
	
	// 存在しないコマンドのテスト。
	err := runCommandE2E("nonexistentcommand12345")
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
	
	// 無効な引数でのコマンドのテスト。
	if IsWindowsTestEnvironment() {
		err = runCommandE2E("dir /invalidarg")
	} else {
		err = runCommandE2E("ls --invalidarg")
	}
	if err == nil {
		t.Error("Expected error for invalid command arguments")
	}
	
	t.Log("✓ Command error handling successful")
}

// =============================================================================
// Step Functions for TestE2ERealRepositoryWorkflow
// =============================================================================

func testStep_CreateRealRepositories(t *testing.T) (string, string) {
	t.Log("Creating real Git repositories...")
	
	tempDir := t.TempDir()
	devRepo := filepath.Join(tempDir, "real-dev")
	opsRepo := filepath.Join(tempDir, "real-ops")
	
	// Dev repository作成。
	err := os.MkdirAll(devRepo, 0755)
	if err != nil {
		t.Fatalf("Failed to create dev repo directory: %v", err)
	}
	
	err = runCommandInDirE2E("git init", devRepo)
	if err != nil {
		t.Fatalf("Failed to init dev repo: %v", err)
	}
	
	err = runCommandInDirE2E("git config user.name TestUser", devRepo)
	if err != nil {
		t.Fatalf("Failed to config dev repo: %v", err)
	}
	
	err = runCommandInDirE2E("git config user.email test@example.com", devRepo)
	if err != nil {
		t.Fatalf("Failed to config dev repo email: %v", err)
	}
	
	// 初期ファイル作成。
	initFile := filepath.Join(devRepo, "README.md")
	err = os.WriteFile(initFile, []byte("# Real Test Repository\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	
	err = runCommandInDirE2E("git add README.md", devRepo)
	if err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m \"feat: initial commit\"", devRepo)
	if err != nil {
		t.Fatalf("Failed to commit initial file: %v", err)
	}
	
	// Ops repository作成（cloneRepositorySimpleを使用）。
	err = cloneRepositorySimpleE2E(devRepo, opsRepo)
	if err != nil {
		t.Fatalf("Failed to clone ops repo: %v", err)
	}
	
	t.Log("✓ Real repositories created successfully")
	t.Logf("  Dev repo: %s", devRepo)
	t.Logf("  Ops repo: %s", opsRepo)
	
	return devRepo, opsRepo
}

func testStep_CleanupRealRepositories(t *testing.T, devRepo, opsRepo string) {
	t.Log("Cleaning up real repositories...")
	
	if err := os.RemoveAll(devRepo); err != nil {
		t.Logf("Warning: Failed to cleanup dev repo: %v", err)
	}
	
	if err := os.RemoveAll(opsRepo); err != nil {
		t.Logf("Warning: Failed to cleanup ops repo: %v", err)
	}
	
	t.Log("✓ Repository cleanup completed")
}

func testStep_CreateRealConfig(t *testing.T, devRepo, opsRepo string) *config.Config {
	t.Log("Creating real configuration...")
	
	cfg := &config.Config{
		DevRepoPath:       devRepo,
		OpsRepoPath:       opsRepo,
		SyncInterval:      "2s",
		FixupInterval:     "5s",
		IncludeExtensions: []string{".go", ".txt", ".md", ".cpp", ".h"},
		ExcludePatterns:   []string{"test/**", "*.tmp", ".git/**"},
		PauseLockFile:     ".sync-paused", // ロックファイル設定を追加
		AutosquashEnabled: false, // テスト環境では無効
		GitExecutable:     "git",
		LogLevel:          "INFO",
		Verbose:           true,
		CommitTemplate:    "Auto-sync: ${timestamp}",
		AuthorName:        "TestUser",
		AuthorEmail:       "test@example.com",
	}
	
	t.Log("✓ Real configuration created")
	return cfg
}

func testStep_RealSyncOperation(t *testing.T, cfg *config.Config) {
	t.Log("Testing real sync operation...")
	
	// Dev側にファイルを追加。
	testFile := filepath.Join(cfg.DevRepoPath, "sync-test.txt")
	err := os.WriteFile(testFile, []byte("sync test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sync test file: %v", err)
	}
	
	err = runCommandInDirE2E("git add sync-test.txt", cfg.DevRepoPath)
	if err != nil {
		t.Fatalf("Failed to add sync test file: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m \"feat: add sync test file\"", cfg.DevRepoPath)
	if err != nil {
		t.Fatalf("Failed to commit sync test file: %v", err)
	}
	
	// Sync実行。
	syncMgr := sync.NewFileSyncer(cfg)
	result, err := syncMgr.Sync()
	if err != nil {
		t.Fatalf("Sync operation failed: %v", err)
	}
	
	t.Logf("  Sync result: Added=%d, Modified=%d, Deleted=%d", 
		len(result.FilesAdded), len(result.FilesModified), len(result.FilesDeleted))
	
	// Ops側にファイルが同期されたか確認。
	opsTestFile := filepath.Join(cfg.OpsRepoPath, "sync-test.txt")
	if _, err := os.Stat(opsTestFile); os.IsNotExist(err) {
		t.Error("Sync test file was not synced to ops repository")
	}
	
	t.Log("✓ Real sync operation successful")
}

func testStep_RealFixupOperation(t *testing.T, cfg *config.Config) {
	t.Log("Testing real fixup operation...")
	
	// Ops側にファイルを変更。
	testFile := filepath.Join(cfg.OpsRepoPath, "fixup-test.txt")
	err := os.WriteFile(testFile, []byte("fixup test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create fixup test file: %v", err)
	}
	
	err = runCommandInDirE2E("git add fixup-test.txt", cfg.OpsRepoPath)
	if err != nil {
		t.Fatalf("Failed to add fixup test file: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m \"feat: add fixup test file\"", cfg.OpsRepoPath)
	if err != nil {
		t.Fatalf("Failed to commit fixup test file: %v", err)
	}
	
	// ファイルを変更（未コミット状態）。
	err = os.WriteFile(testFile, []byte("modified fixup test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify fixup test file: %v", err)
	}
	
	// Fixup実行。
	fixupMgr := fixup.NewFixupManager(cfg)
	result, err := fixupMgr.RunFixup()
	if err != nil {
		t.Fatalf("Fixup operation failed: %v", err)
	}
	
	t.Logf("  Fixup result: Modified=%d files, Success=%v", 
		result.FilesModified, result.Success)
	
	t.Log("✓ Real fixup operation successful")
}

func testStep_RealIntegratedWorkflow(t *testing.T, cfg *config.Config, devRepo, opsRepo string) {
	t.Log("Testing real integrated workflow...")
	
	// 開発者がDev側でファイルを追加。
	for i := 1; i <= 3; i++ {
		filename := fmt.Sprintf("workflow-test-%d.txt", i)
		testFile := filepath.Join(devRepo, filename)
		content := fmt.Sprintf("Workflow test content %d", i)
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create workflow test file %d: %v", i, err)
		}
		
		err = runCommandInDirE2E(fmt.Sprintf("git add %s", filename), devRepo)
		if err != nil {
			t.Fatalf("Failed to add workflow test file %d: %v", i, err)
		}
		
		err = runCommandInDirE2E(fmt.Sprintf("git commit -m \"feat: add workflow test %d\"", i), devRepo)
		if err != nil {
			t.Fatalf("Failed to commit workflow test file %d: %v", i, err)
		}
	}
	
	// Sync実行。
	syncMgr := sync.NewFileSyncer(cfg)
	syncResult, err := syncMgr.Sync()
	if err != nil {
		t.Fatalf("Integrated sync failed: %v", err)
	}
	
	t.Logf("  Integrated sync result: %+v", syncResult)
	
	// Ops側での変更作業。
	opsFile := filepath.Join(opsRepo, "ops-specific.txt")
	err = os.WriteFile(opsFile, []byte("ops-specific content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create ops-specific file: %v", err)
	}
	
	err = runCommandInDirE2E("git add ops-specific.txt", opsRepo)
	if err != nil {
		t.Fatalf("Failed to add ops-specific file: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m \"feat: add ops-specific file\"", opsRepo)
	if err != nil {
		t.Fatalf("Failed to commit ops-specific file: %v", err)
	}
	
	// 変更してFixup実行。
	err = os.WriteFile(opsFile, []byte("modified ops-specific content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify ops-specific file: %v", err)
	}
	
	fixupMgr := fixup.NewFixupManager(cfg)
	fixupResult, err := fixupMgr.RunFixup()
	if err != nil {
		t.Fatalf("Integrated fixup failed: %v", err)
	}
	
	t.Logf("  Integrated fixup result: %+v", fixupResult)
	
	t.Log("✓ Real integrated workflow successful")
}

// =============================================================================
// Step Functions for TestE2ECompleteWorkflow
// =============================================================================

func testStep_CompleteEnvironmentSetup(t *testing.T) (string, string) {
	t.Log("Setting up complete test environment...")
	
	testDir := SetupTestEnvironment(t)
	
	devRepo := filepath.Join(testDir, "repos", "complete-dev")
	opsRepo := filepath.Join(testDir, "repos", "complete-ops")
	
	// 既存ディレクトリクリーンアップ。
	os.RemoveAll(devRepo)
	os.RemoveAll(opsRepo)
	
	// Dev repository作成。
	err := os.MkdirAll(devRepo, 0755)
	if err != nil {
		t.Fatalf("Failed to create complete dev repo: %v", err)
	}
	
	err = runCommandInDirE2E("git init", devRepo)
	if err != nil {
		t.Fatalf("Failed to init complete dev repo: %v", err)
	}
	
	err = runCommandInDirE2E("git config user.name CompleteTestUser", devRepo)
	if err != nil {
		t.Fatalf("Failed to config complete dev repo: %v", err)
	}
	
	err = runCommandInDirE2E("git config user.email complete@example.com", devRepo)
	if err != nil {
		t.Fatalf("Failed to config complete dev repo email: %v", err)
	}
	
	// 初期構造作成。
	dirs := []string{"src", "docs", "tests"}
	for _, dir := range dirs {
		dirPath := filepath.Join(devRepo, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		// 各ディレクトリに初期ファイル作成。
		initFile := filepath.Join(dirPath, "README.md")
		content := fmt.Sprintf("# %s Directory\n\nThis is the %s directory.\n", dir, dir)
		if err := os.WriteFile(initFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s README: %v", dir, err)
		}
	}
	
	// 初期コミット。
	err = runCommandInDirE2E("git add .", devRepo)
	if err != nil {
		t.Fatalf("Failed to add initial structure: %v", err)
	}
	
	err = runCommandInDirE2E("git commit -m \"feat: initial project structure\"", devRepo)
	if err != nil {
		t.Fatalf("Failed to commit initial structure: %v", err)
	}
	
	// Ops repository作成。
	err = cloneRepositorySimpleE2E(devRepo, opsRepo)
	if err != nil {
		t.Fatalf("Failed to clone complete ops repo: %v", err)
	}
	
	t.Log("✓ Complete environment setup successful")
	return devRepo, opsRepo
}

func testStep_CompleteCleanup(t *testing.T, devRepo, opsRepo string) {
	t.Log("Performing complete cleanup...")
	
	if err := os.RemoveAll(devRepo); err != nil {
		t.Logf("Warning: Failed to cleanup complete dev repo: %v", err)
	}
	
	if err := os.RemoveAll(opsRepo); err != nil {
		t.Logf("Warning: Failed to cleanup complete ops repo: %v", err)
	}
	
	t.Log("✓ Complete cleanup successful")
}

func testStep_InitializeManagers(t *testing.T, devRepo, opsRepo string) (*config.Config, *sync.FileSyncer, *fixup.FixupManager) {
	t.Log("Initializing managers...")
	
	cfg := &config.Config{
		DevRepoPath:       devRepo,
		OpsRepoPath:       opsRepo,
		SyncInterval:      "3s",
		FixupInterval:     "7s",
		IncludeExtensions: []string{".go", ".txt", ".md", ".cpp", ".h", ".py", ".js"},
		ExcludePatterns:   []string{".git/**", "*.tmp", "*.log"},
		PauseLockFile:     ".sync-paused", // ロックファイル設定を追加
		AutosquashEnabled: false,
		GitExecutable:     "git",
		LogLevel:          "INFO",
		Verbose:           true,
		CommitTemplate:    "Auto-sync: Complete workflow test at ${timestamp}",
		FixupMsgPrefix:    "fixup! ",
		AuthorName:        "CompleteTestUser",
		AuthorEmail:       "complete@example.com",
	}
	
	syncMgr := sync.NewFileSyncer(cfg)
	fixupMgr := fixup.NewFixupManager(cfg)
	
	t.Log("✓ Managers initialized successfully")
	return cfg, syncMgr, fixupMgr
}

func testStep_SimulateContinuousDevelopment(t *testing.T, ctx context.Context, devRepo string, syncMgr *sync.FileSyncer, fixupMgr *fixup.FixupManager) {
	t.Log("Simulating continuous development...")
	
	// 開発シミュレーション用goroutine。
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		
		fileCount := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fileCount++
				
				// ファイル作成。
				filename := fmt.Sprintf("auto-file-%d.txt", fileCount)
				filepath := filepath.Join(devRepo, "src", filename)
				content := fmt.Sprintf("Auto-generated content %d at %v", fileCount, time.Now().Format("15:04:05"))
				
				if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
					t.Logf("Warning: Failed to create auto file %d: %v", fileCount, err)
					continue
				}
				
				if err := runCommandInDirE2E(fmt.Sprintf("git add src/%s", filename), devRepo); err != nil {
					t.Logf("Warning: Failed to add auto file %d: %v", fileCount, err)
					continue
				}
				
				if err := runCommandInDirE2E(fmt.Sprintf("git commit -m \"feat: auto-add file %d\"", fileCount), devRepo); err != nil {
					t.Logf("Warning: Failed to commit auto file %d: %v", fileCount, err)
					continue
				}
				
				t.Logf("  Created and committed: %s", filename)
			}
		}
	}()
	
	// Sync シミュレーション用goroutine。
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result, err := syncMgr.Sync()
				if err != nil {
					t.Logf("Sync error: %v", err)
				} else if len(result.FilesAdded)+len(result.FilesModified)+len(result.FilesDeleted) > 0 {
					t.Logf("  Sync result: +%d ~%d -%d", 
						len(result.FilesAdded), len(result.FilesModified), len(result.FilesDeleted))
				}
			}
		}
	}()
	
	// Fixup シミュレーション用goroutine。
	go func() {
		ticker := time.NewTicker(7 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result, err := fixupMgr.RunFixup()
				if err != nil {
					t.Logf("Fixup error: %v", err)
				} else if result.FilesModified > 0 {
					t.Logf("  Fixup result: %d files modified", result.FilesModified)
				}
			}
		}
	}()
	
	// 統合動作を監視。
	<-ctx.Done()
	t.Log("✓ Continuous development simulation completed")
}

func testStep_VerifyWorkflowResults(t *testing.T, cfg *config.Config, devRepo, opsRepo string) {
	t.Log("Verifying workflow results...")
	
	// Dev repositoryのコミット数確認。
	devCommitCount, err := getCommitCountE2E(devRepo)
	if err != nil {
		t.Logf("Warning: Failed to get dev commit count: %v", err)
	} else {
		t.Logf("  Dev repository commits: %d", devCommitCount)
		if devCommitCount <= 1 {
			t.Error("Expected more than 1 commit in dev repository")
		}
	}
	
	// Ops repositoryのコミット数確認。
	opsCommitCount, err := getCommitCountE2E(opsRepo)
	if err != nil {
		t.Logf("Warning: Failed to get ops commit count: %v", err)
	} else {
		t.Logf("  Ops repository commits: %d", opsCommitCount)
	}
	
	// ファイル同期確認。
	devFiles, err := countFilesInDir(filepath.Join(devRepo, "src"))
	if err != nil {
		t.Logf("Warning: Failed to count dev files: %v", err)
	} else {
		t.Logf("  Dev src files: %d", devFiles)
	}
	
	opsFiles, err := countFilesInDir(filepath.Join(opsRepo, "src"))
	if err != nil {
		t.Logf("Warning: Failed to count ops files: %v", err)
	} else {
		t.Logf("  Ops src files: %d", opsFiles)
	}
	
	// 基本的な同期確認。
	if devFiles > 0 && opsFiles == 0 {
		t.Error("Files were created in dev but not synced to ops")
	}
	
	t.Log("✓ Workflow results verification completed")
}

// =============================================================================
// Helper Functions
// =============================================================================

func runCommandE2E(command string) error {
	// runCommand関数の実装をここで模倣（実際のcmd/run.goのrunCommandを呼び出せない場合）
	var cmd *exec.Cmd
	
	if IsWindowsTestEnvironment() {
		cmd = exec.Command("cmd", "/C", command)
		cmd.Dir = "C:\\"
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution error: %w, output: %s", err, string(output))
	}
	
	return nil
}

func runCommandInDirE2E(command, dir string) error {
	// runCommandInDir関数の実装をここで模倣
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	var cmd *exec.Cmd
	
	if IsWindowsTestEnvironment() {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	
	cmd.Dir = dir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command execution error: %w, output: %s", err, string(output))
	}
	
	return nil
}

func cloneRepositorySimpleE2E(srcPath, destPath string) error {
	// cloneRepositorySimple関数の実装をここで模倣
	parentDir := filepath.Dir(destPath)
	
	if parentDir != "." && !isWindowsDriveRoot(parentDir) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("directory creation error: %v", err)
		}
	}
	
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination directory already exists: %s", destPath)
	}
	
	cmd := fmt.Sprintf("git clone %s %s", srcPath, destPath)
	return runCommandE2E(cmd)
}

func getCommitCountE2E(repoPath string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count); err != nil {
		return 0, err
	}
	
	return count, nil
}

func countFilesInDir(dirPath string) (int, error) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return 0, nil
	}
	
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, file := range files {
		if !file.IsDir() {
			count++
		}
	}
	
	return count, nil
}

func isWindowsDriveRoot(path string) bool {
	return len(path) == 2 && path[1] == ':'
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}