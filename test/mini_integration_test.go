package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMini1_BasicEnvironment は最小限の環境テスト（TDD Step 1: Red -> Green -> Refactor）
func TestMini1_BasicEnvironment(t *testing.T) {
	t.Log("Mini Test 1: Basic environment setup")
	
	// テストディレクトリが作成できることを確認
	testDir := "../test"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		// まだ存在しない場合は作成
		err = os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}
	
	t.Log("✓ Basic environment OK")
}

// TestMini2_GitAvailability はGitコマンドの利用可能性をテスト（TDD Step 2）
func TestMini2_GitAvailability(t *testing.T) {
	t.Log("Mini Test 2: Git command availability")
	
	err := runGitCommand("", "version")
	if err != nil {
		t.Fatalf("Git command not available: %v", err)
	}
	
	t.Log("✓ Git command available")
}

// TestMini3_CreateSimpleRepo は最小限のリポジトリ作成をテスト（TDD Step 3）
func TestMini3_CreateSimpleRepo(t *testing.T) {
	t.Log("Mini Test 3: Create simple repository")
	
	repoPath := filepath.Join("test", "repos", "mini-test-repo")
	
	// クリーンアップ
	os.RemoveAll(repoPath)
	
	// リポジトリ作成
	err := createTestRepository(t, repoPath)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	
	// 作成確認
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatal("Git repository was not created properly")
	}
	
	t.Log("✓ Simple repository created")
}

// TestMini4_BasicFileOperation は基本的なファイル操作をテスト（TDD Step 4）
func TestMini4_BasicFileOperation(t *testing.T) {
	t.Log("Mini Test 4: Basic file operations")
	
	repoPath := filepath.Join("test", "repos", "mini-test-repo")
	
	// テストファイル作成
	testFile := filepath.Join(repoPath, "test-file.txt")
	content := "Hello, TDD!"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// ファイル読み込み確認
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	if string(readContent) != content {
		t.Fatalf("File content mismatch: expected %s, got %s", content, string(readContent))
	}
	
	t.Log("✓ Basic file operations OK")
}

// TestMini5_GitCommitOperation は基本的なGitコミット操作をテスト（TDD Step 5）
func TestMini5_GitCommitOperation(t *testing.T) {
	t.Log("Mini Test 5: Git commit operations")
	
	repoPath := filepath.Join("test", "repos", "mini-test-repo")
	
	// テストファイル作成
	testFile := "mini-test.txt"
	fullPath := filepath.Join(repoPath, testFile)
	content := "TDD test content"
	
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Git add
	err = gitAdd(repoPath, testFile)
	if err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}
	
	// Git commit
	err = gitCommit(repoPath, "feat: add mini test file")
	if err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
	
	// コミット数確認
	count, err := getCommitCount(repoPath)
	if err != nil {
		t.Fatalf("Failed to get commit count: %v", err)
	}
	
	if count < 2 { // 初期コミット + テストコミット
		t.Fatalf("Expected at least 2 commits, got %d", count)
	}
	
	t.Log("✓ Git commit operations OK")
}

// TestMini6_WindowsEnvironmentCheck はWindows環境の基本チェック（TDD Step 6）
func TestMini6_WindowsEnvironmentCheck(t *testing.T) {
	t.Log("Mini Test 6: Windows environment check")
	
	if !IsWindowsTestEnvironment() {
		t.Skip("Skipping Windows-specific test")
	}
	
	// PowerShellの利用可能性確認
	err := runPowerShellCommand("Write-Host 'PowerShell OK'")
	if err != nil {
		t.Fatalf("PowerShell not available: %v", err)
	}
	
	t.Log("✓ Windows environment check OK")
}

// ヘルパー関数（最小限）
func runPowerShellCommand(command string) error {
	cmd := exec.Command("powershell", "-Command", command)
	_, err := cmd.CombinedOutput()
	return err
}