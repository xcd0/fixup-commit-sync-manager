package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/vhdx"
)

// TestMini7_VHDXManagerCreation はVHDXマネージャー作成をテスト（TDD Step 7）
func TestMini7_VHDXManagerCreation(t *testing.T) {
	t.Log("Mini Test 7: VHDX Manager creation")
	
	vhdxPath := filepath.Join("..", "test", "vhdx", "mini-test.vhdx")
	manager := vhdx.NewVHDXManager(vhdxPath, "T:", "1GB", false)
	
	if manager == nil {
		t.Fatal("Failed to create VHDX manager")
	}
	
	if manager.VHDXPath != vhdxPath {
		t.Fatalf("VHDX path mismatch: expected %s, got %s", vhdxPath, manager.VHDXPath)
	}
	
	t.Log("✓ VHDX Manager creation OK")
}

// TestMini8_VHDXDirectorySetup はVHDXディレクトリ準備をテスト（TDD Step 8）
func TestMini8_VHDXDirectorySetup(t *testing.T) {
	t.Log("Mini Test 8: VHDX directory setup")
	
	vhdxDir := filepath.Join("..", "test", "vhdx")
	err := os.MkdirAll(vhdxDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create VHDX directory: %v", err)
	}
	
	// ディレクトリの存在確認
	if _, err := os.Stat(vhdxDir); os.IsNotExist(err) {
		t.Fatal("VHDX directory was not created")
	}
	
	t.Log("✓ VHDX directory setup OK")
}

// TestMini9_VHDXCreateDryRun はVHDX作成のドライランをテスト（TDD Step 9）
func TestMini9_VHDXCreateDryRun(t *testing.T) {
	t.Log("Mini Test 9: VHDX create dry run")
	
	vhdxPath := filepath.Join("..", "test", "vhdx", "dryrun-test.vhdx")
	
	// 既存ファイルをクリーンアップ
	os.Remove(vhdxPath)
	
	manager := vhdx.NewVHDXManager(vhdxPath, "T:", "1GB", false)
	
	// VHDXディレクトリが存在することを確認
	vhdxDir := filepath.Dir(vhdxPath)
	if _, err := os.Stat(vhdxDir); os.IsNotExist(err) {
		t.Fatalf("VHDX directory does not exist: %s", vhdxDir)
	}
	
	// 実際の作成はWindows環境でのみテスト
	if !IsWindowsTestEnvironment() {
		t.Log("  Skipping actual VHDX creation on non-Windows")
		t.Log("✓ VHDX create dry run OK (skipped)")
		return
	}
	
	// Windows環境での実際の作成テスト
	err := manager.Create("1GB", false)
	if err != nil {
		// 権限やファイルシステムの制限で失敗する場合は警告として記録
		if containsAny(err.Error(), []string{"Access is denied", "file system limitation", "administrator"}) {
			t.Logf("  VHDX creation skipped due to permissions: %v", err)
			t.Log("✓ VHDX create dry run OK (permission limited)")
			return
		}
		t.Fatalf("VHDX creation failed: %v", err)
	}
	
	// 作成成功の場合はクリーンアップ
	defer func() {
		if err := os.Remove(vhdxPath); err != nil {
			t.Logf("Warning: Failed to cleanup VHDX file: %v", err)
		}
	}()
	
	// VHDXファイルの存在確認
	if _, err := os.Stat(vhdxPath); os.IsNotExist(err) {
		t.Fatal("VHDX file was not created")
	}
	
	t.Log("✓ VHDX create dry run OK (created)")
}

// TestMini10_RepositoryCloneBasic は基本的なリポジトリクローンをテスト（TDD Step 10）
func TestMini10_RepositoryCloneBasic(t *testing.T) {
	t.Log("Mini Test 10: Repository clone basic")
	
	// ソースリポジトリパス
	srcRepo := filepath.Join("test", "repos", "mini-test-repo")
	
	// ソースリポジトリの存在確認
	if _, err := os.Stat(srcRepo); os.IsNotExist(err) {
		t.Logf("Source repository not found at %s, creating it...", srcRepo)
		// ソースリポジトリが存在しない場合は作成
		err := createTestRepository(t, srcRepo)
		if err != nil {
			t.Fatalf("Failed to create source repository: %v", err)
		}
	}
	
	// クローン先パス
	cloneRepo := filepath.Join("test", "repos", "mini-clone-repo")
	
	// 既存クローンのクリーンアップ
	os.RemoveAll(cloneRepo)
	
	// クローン実行
	err := cloneTestRepository(t, srcRepo, cloneRepo)
	if err != nil {
		t.Fatalf("Repository clone failed: %v", err)
	}
	
	// クローンしたリポジトリの確認
	gitDir := filepath.Join(cloneRepo, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatal("Cloned repository .git directory not found")
	}
	
	// READMEファイルの確認
	readmeFile := filepath.Join(cloneRepo, "README.md")
	if _, err := os.Stat(readmeFile); os.IsNotExist(err) {
		t.Fatal("README.md not found in cloned repository")
	}
	
	t.Log("✓ Repository clone basic OK")
}

// TestMini11_ConfigBasic は基本的な設定作成をテスト（TDD Step 11）
func TestMini11_ConfigBasic(t *testing.T) {
	t.Log("Mini Test 11: Config basic")
	
	devRepo := filepath.Join("test", "repos", "mini-test-repo")
	opsRepo := filepath.Join("test", "repos", "mini-clone-repo")
	
	// 最小限の設定作成
	cfg := createMiniConfig(devRepo, opsRepo)
	
	// 設定値の確認
	if cfg.DevRepoPath != devRepo {
		t.Fatalf("Dev repo path mismatch: expected %s, got %s", devRepo, cfg.DevRepoPath)
	}
	
	if cfg.OpsRepoPath != opsRepo {
		t.Fatalf("Ops repo path mismatch: expected %s, got %s", opsRepo, cfg.OpsRepoPath)
	}
	
	if cfg.SyncInterval != "5s" {
		t.Fatalf("Sync interval mismatch: expected 5s, got %s", cfg.SyncInterval)
	}
	
	if cfg.FixupInterval != "20s" {
		t.Fatalf("Fixup interval mismatch: expected 20s, got %s", cfg.FixupInterval)
	}
	
	t.Log("✓ Config basic OK")
}

// ヘルパー関数

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(str, substr) {
			return true
		}
	}
	return false
}

func createMiniConfig(devRepo, opsRepo string) *config.Config {
	return &config.Config{
		DevRepoPath:       devRepo,
		OpsRepoPath:       opsRepo,
		SyncInterval:      "5s",
		FixupInterval:     "20s",
		IncludeExtensions: []string{".go", ".txt", ".md"},
		ExcludePatterns:   []string{"test/**", "*.tmp"},
		AutosquashEnabled: true,
		GitExecutable:     "git",
		LogLevel:          "DEBUG",
		Verbose:           true,
	}
}