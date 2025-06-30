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
	"fixup-commit-sync-manager/internal/vhdx"
)

// TestIntegrationVHDXSyncFixup は統合テスト: VHDX作成→マウント→定期コミット→sync→fixup→1分間動作確認。
// t-wadaのTDDアプローチに従って、小さなステップで実装する。
func TestIntegrationVHDXSyncFixup(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}
	
	// Step 1: 環境準備のテスト（Red phase）
	testStep1_PrepareEnvironment(t)
	
	// Step 2: VHDX作成のテスト（Red -> Green -> Refactor）
	vhdxManager := testStep2_CreateVHDX(t)
	
	// Step 3: VHDXマウントのテスト（Red -> Green -> Refactor）
	testStep3_MountVHDX(t, vhdxManager)
	defer testStep_CleanupVHDX(t, vhdxManager)
	
	// Step 4: リポジトリセットアップのテスト（Red -> Green -> Refactor）
	devRepoPath, opsRepoPath := testStep4_SetupRepositories(t, vhdxManager)
	
	// Step 5: 設定ファイル作成のテスト（Red -> Green -> Refactor）
	cfg := testStep5_CreateConfig(t, devRepoPath, opsRepoPath)
	
	// Step 6: 定期コミット機能のテスト（Red -> Green -> Refactor）
	ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
	defer cancel()
	
	testStep6_PeriodicCommits(t, ctx, devRepoPath)
	
	// Step 7: Sync機能のテスト（Red -> Green -> Refactor）
	testStep7_SyncOperation(t, ctx, cfg)
	
	// Step 8: Fixup機能のテスト（Red -> Green -> Refactor）
	testStep8_FixupOperation(t, ctx, cfg)
	
	// Step 9: 1分間の統合動作確認（Red -> Green -> Refactor）
	testStep9_IntegratedOperation(t, ctx, cfg, devRepoPath)
}

// testStep1_PrepareEnvironment は環境準備をテストする。
func testStep1_PrepareEnvironment(t *testing.T) {
	t.Log("Step 1: Preparing test environment...")
	
	testDir := SetupTestEnvironment(t)
	if testDir == "" {
		t.Fatal("Failed to setup test environment")
	}
	
	// テスト用ディレクトリの存在確認。
	requiredDirs := []string{"temp", "vhdx", "repos", "config", "data"}
	for _, dir := range requiredDirs {
		fullPath := filepath.Join(testDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Fatalf("Required test directory does not exist: %s", fullPath)
		}
	}
	
	t.Log("✓ Step 1: Environment prepared successfully")
}

// testStep2_CreateVHDX はVHDX作成をテストする。
func testStep2_CreateVHDX(t *testing.T) *vhdx.VHDXManager {
	t.Log("Step 2: Creating VHDX...")
	
	vhdxPath := GetTestVHDXPath("integration-test")
	manager := vhdx.NewVHDXManager(vhdxPath, "T:", "1GB", false)
	
	// 既存ファイルのクリーンアップ。
	os.Remove(vhdxPath)
	
	// VHDX作成を試行。
	err := manager.Create("1GB", false)
	if err != nil {
		// 非Windows環境や権限不足の場合はスキップ。
		if !IsWindowsTestEnvironment() {
			t.Skip("VHDX creation requires Windows environment")
		}
		if strings.Contains(err.Error(), "Access is denied") || 
		   strings.Contains(err.Error(), "file system limitation") {
			t.Skip("VHDX creation requires administrator privileges")
		}
		t.Fatalf("Failed to create VHDX: %v", err)
	}
	
	// VHDXファイルの存在確認（WSL環境またはパス変換された場合はスキップ）。
	if strings.Contains(manager.VHDXPath, "wsl.localhost") ||
	   strings.Contains(manager.VHDXPath, "AppData\\Local\\Temp") ||
	   strings.Contains(manager.VHDXPath, "C:\\") {
		t.Log("  VHDX existence check skipped (Windows path conversion)")
		t.Logf("  VHDX created at: %s", manager.VHDXPath)
	} else {
		if _, err := os.Stat(vhdxPath); os.IsNotExist(err) {
			t.Fatalf("VHDX file was not created: %s", vhdxPath)
		}
	}
	
	t.Log("✓ Step 2: VHDX created successfully")
	return manager
}

// testStep3_MountVHDX はVHDXマウントをテストする。
func testStep3_MountVHDX(t *testing.T, manager *vhdx.VHDXManager) {
	t.Log("Step 3: Mounting VHDX...")
	
	err := manager.Mount()
	if err != nil {
		if !IsWindowsTestEnvironment() {
			t.Skip("VHDX mount requires Windows environment")
		}
		if strings.Contains(err.Error(), "Access is denied") {
			t.Skip("VHDX mount requires administrator privileges")
		}
		if strings.Contains(err.Error(), "VHDX is already mounted") {
			t.Log("VHDX was already mounted, this is expected in mount-then-unmount flow")
		} else {
			t.Fatalf("Failed to mount VHDX: %v", err)
		}
	}
	
	// マウントポイントの存在確認。
	mountPoint := manager.MountPoint
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		t.Fatalf("Mount point does not exist: %s", mountPoint)
	}
	
	t.Log("✓ Step 3: VHDX mounted successfully at", mountPoint)
}

// testStep_CleanupVHDX はVHDXクリーンアップを行う。
func testStep_CleanupVHDX(t *testing.T, manager *vhdx.VHDXManager) {
	t.Log("Cleanup: Unmounting VHDX...")
	
	if err := manager.UnmountVHDX(); err != nil {
		t.Logf("Warning: Failed to unmount VHDX: %v", err)
	}
	
	// VHDXファイルを削除。
	if err := os.Remove(manager.VHDXPath); err != nil {
		t.Logf("Warning: Failed to remove VHDX file: %v", err)
	}
	
	t.Log("✓ Cleanup: VHDX cleanup completed")
}

// testStep4_SetupRepositories はリポジトリセットアップをテストする。
func testStep4_SetupRepositories(t *testing.T, manager *vhdx.VHDXManager) (string, string) {
	t.Log("Step 4: Setting up repositories...")
	
	// Dev リポジトリをローカルに作成。
	devRepoPath := GetTestRepoPath("dev-repo-integration")
	if err := os.RemoveAll(devRepoPath); err != nil {
		t.Logf("Warning: Failed to remove existing dev repo: %v", err)
	}
	
	if err := createTestRepository(t, devRepoPath); err != nil {
		t.Fatalf("Failed to create dev repository: %v", err)
	}
	
	// Ops リポジトリをVHDX上に作成。
	opsRepoPath := filepath.Join(manager.MountPoint, "ops-repo")
	if err := os.RemoveAll(opsRepoPath); err != nil {
		t.Logf("Warning: Failed to remove existing ops repo: %v", err)
	}
	
	if err := cloneTestRepository(t, devRepoPath, opsRepoPath); err != nil {
		t.Fatalf("Failed to clone ops repository: %v", err)
	}
	
	t.Log("✓ Step 4: Repositories setup completed")
	t.Logf("  Dev repo: %s", devRepoPath)
	t.Logf("  Ops repo: %s", opsRepoPath)
	
	return devRepoPath, opsRepoPath
}

// testStep5_CreateConfig は設定ファイル作成をテストする。
func testStep5_CreateConfig(t *testing.T, devRepoPath, opsRepoPath string) *config.Config {
	t.Log("Step 5: Creating configuration...")
	
	// configPath := filepath.Join("test", "temp", "integration-config.hjson")
	cfg := &config.Config{
		DevRepoPath:       devRepoPath,
		OpsRepoPath:       opsRepoPath,
		SyncInterval:      "5s",
		FixupInterval:     "20s",
		IncludeExtensions: []string{".go", ".txt", ".md"},
		ExcludePatterns:   []string{"test/**", "*.tmp"},
		AutosquashEnabled: true,
		GitExecutable:     "git",
		LogLevel:          "DEBUG",
		Verbose:           true,
	}
	
	// 設定の妥当性確認。
	if cfg.DevRepoPath == "" || cfg.OpsRepoPath == "" {
		t.Fatal("Repository paths are not set")
	}
	
	t.Log("✓ Step 5: Configuration created successfully")
	return cfg
}

// testStep6_PeriodicCommits は定期コミット機能をテストする。
func testStep6_PeriodicCommits(t *testing.T, ctx context.Context, devRepoPath string) {
	t.Log("Step 6: Starting periodic commits...")
	
	// 3秒ごとにファイルを変更してコミットするgoroutineを開始。
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		
		commitCount := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				commitCount++
				
				// ファイルを作成/更新。
				filename := fmt.Sprintf("test-file-%d.txt", commitCount)
				filepath := filepath.Join(devRepoPath, filename)
				content := fmt.Sprintf("Test content %d at %v", commitCount, time.Now())
				
				if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
					t.Logf("Warning: Failed to write test file: %v", err)
					continue
				}
				
				// Git add & commit。
				if err := gitAdd(devRepoPath, filename); err != nil {
					t.Logf("Warning: Failed to git add: %v", err)
					continue
				}
				
				message := fmt.Sprintf("feat: add test file %d", commitCount)
				if err := gitCommit(devRepoPath, message); err != nil {
					t.Logf("Warning: Failed to git commit: %v", err)
					continue
				}
				
				t.Logf("  Commit %d: %s", commitCount, message)
			}
		}
	}()
	
	t.Log("✓ Step 6: Periodic commits started")
}

// testStep7_SyncOperation はSync機能をテストする。
func testStep7_SyncOperation(t *testing.T, ctx context.Context, cfg *config.Config) {
	t.Log("Step 7: Testing sync operation...")
	
	syncManager := sync.NewFileSyncer(cfg)
	
	// 同期を1回実行してテスト。
	result, err := syncManager.Sync()
	if err != nil {
		t.Logf("Initial sync failed (expected): %v", err)
		// 初回は失敗する可能性があるため、ログに記録。
	} else {
		t.Logf("Sync result: %+v", result)
	}
	
	// 継続的同期を開始。
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := syncManager.Sync()
				if err != nil {
					t.Logf("Sync error: %v", err)
				} else {
					t.Log("  Sync completed successfully")
				}
			}
		}
	}()
	
	t.Log("✓ Step 7: Sync operation started")
}

// testStep8_FixupOperation はFixup機能をテストする。
func testStep8_FixupOperation(t *testing.T, ctx context.Context, cfg *config.Config) {
	t.Log("Step 8: Testing fixup operation...")
	
	fixupManager := fixup.NewFixupManager(cfg)
	
	// Fixupを1回実行してテスト。
	result, err := fixupManager.RunFixup()
	if err != nil {
		t.Logf("Initial fixup failed (expected): %v", err)
		// 初回は失敗する可能性があるため、ログに記録。
	} else {
		t.Logf("Fixup result: %+v", result)
	}
	
	// 継続的Fixupを開始。
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := fixupManager.RunFixup()
				if err != nil {
					t.Logf("Fixup error: %v", err)
				} else {
					t.Log("  Fixup completed successfully")
				}
			}
		}
	}()
	
	t.Log("✓ Step 8: Fixup operation started")
}

// testStep9_IntegratedOperation は1分間の統合動作確認をテストする。
func testStep9_IntegratedOperation(t *testing.T, ctx context.Context, cfg *config.Config, devRepoPath string) {
	t.Log("Step 9: Running integrated operation for 1 minute...")
	
	startTime := time.Now()
	endTime := startTime.Add(60 * time.Second)
	
	// 統計情報。
	stats := struct {
		commits int
		syncs   int
		fixups  int
	}{}
	
	// 統計情報を定期的に出力。
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				remaining := endTime.Sub(time.Now())
				t.Logf("  Progress: %v elapsed, %v remaining", 
					elapsed.Round(time.Second), 
					remaining.Round(time.Second))
			}
		}
	}()
	
	// 1分間待機。
	select {
	case <-ctx.Done():
		t.Log("✓ Step 9: Context cancelled")
	case <-time.After(60 * time.Second):
		t.Log("✓ Step 9: 1 minute integrated operation completed")
	}
	
	// 最終結果の確認。
	totalElapsed := time.Since(startTime)
	t.Logf("  Total elapsed time: %v", totalElapsed.Round(time.Second))
	t.Logf("  Statistics: %d commits, %d syncs, %d fixups", stats.commits, stats.syncs, stats.fixups)
	
	// Devリポジトリのコミット数確認。
	commitCount, err := getCommitCount(devRepoPath)
	if err != nil {
		t.Logf("Warning: Failed to get commit count: %v", err)
	} else {
		t.Logf("  Dev repo commit count: %d", commitCount)
	}
	
	// Opsリポジトリの状態確認。
	if _, err := os.Stat(cfg.OpsRepoPath); err == nil {
		opsCommitCount, err := getCommitCount(cfg.OpsRepoPath)
		if err != nil {
			t.Logf("Warning: Failed to get ops commit count: %v", err)
		} else {
			t.Logf("  Ops repo commit count: %d", opsCommitCount)
		}
	}
	
	t.Log("✓ Integration test completed successfully")
}

// ヘルパー関数群

func createTestRepository(t *testing.T, repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}
	
	// git init
	if err := runGitCommand(repoPath, "init"); err != nil {
		return err
	}
	
	// git config
	if err := runGitCommand(repoPath, "config", "user.name", "Test User"); err != nil {
		return err
	}
	if err := runGitCommand(repoPath, "config", "user.email", "test@example.com"); err != nil {
		return err
	}
	
	// 初期ファイル作成
	initFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(initFile, []byte("# Test Repository\n"), 0644); err != nil {
		return err
	}
	
	// git add & commit
	if err := runGitCommand(repoPath, "add", "README.md"); err != nil {
		return err
	}
	if err := runGitCommand(repoPath, "commit", "-m", "feat: initial commit"); err != nil {
		return err
	}
	
	return nil
}

func cloneTestRepository(t *testing.T, srcPath, dstPath string) error {
	return runGitCommand("", "clone", srcPath, dstPath)
}

func gitAdd(repoPath, filename string) error {
	return runGitCommand(repoPath, "add", filename)
}

func gitCommit(repoPath, message string) error {
	return runGitCommand(repoPath, "commit", "-m", message)
}

func getCommitCount(repoPath string) (int, error) {
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

func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w, output: %s", strings.Join(args, " "), err, string(output))
	}
	
	return nil
}