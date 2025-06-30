package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// SetupTestEnvironment はテスト環境をセットアップする。
func SetupTestEnvironment(t *testing.T) string {
	// プロジェクトルートを取得。
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	
	// テストディレクトリを作成。
	testDir := filepath.Join(projectRoot, "test")
	
	// 必要なサブディレクトリを作成。
	dirs := []string{
		filepath.Join(testDir, "temp"),
		filepath.Join(testDir, "vhdx"),
		filepath.Join(testDir, "repos"),
		filepath.Join(testDir, "config"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}
	
	return testDir
}

// CleanupTestFiles はテストファイルをクリーンアップする。
func CleanupTestFiles(t *testing.T, patterns ...string) {
	// プロジェクトルートを取得。
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	testDir := filepath.Join(projectRoot, "test")
	
	// デフォルトのクリーンアップパターン。
	if len(patterns) == 0 {
		patterns = []string{
			"temp/*",
			"vhdx/*.vhdx",
			"vhdx/*.vhd",
			"repos/test-*",
		}
	}
	
	for _, pattern := range patterns {
		fullPattern := filepath.Join(testDir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			t.Logf("Failed to glob pattern %s: %v", fullPattern, err)
			continue
		}
		
		for _, match := range matches {
			if err := os.RemoveAll(match); err != nil {
				t.Logf("Failed to remove %s: %v", match, err)
			}
		}
	}
}

// IsWindowsTestEnvironment はWindows環境でのテスト実行かどうかを判定する。
func IsWindowsTestEnvironment() bool {
	return runtime.GOOS == "windows" || os.Getenv("GOOS") == "windows"
}

// SkipIfNotWindows はWindows環境でない場合にテストをスキップする。
func SkipIfNotWindows(t *testing.T) {
	if !IsWindowsTestEnvironment() {
		t.Skip("This test requires Windows environment or GOOS=windows")
	}
}

// GetTestVHDXPath はテスト用VHDXファイルのパスを返す。
func GetTestVHDXPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	return filepath.Join(projectRoot, "test", "vhdx", name+".vhdx")
}

// GetTestRepoPath はテスト用リポジトリのパスを返す。
func GetTestRepoPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	return filepath.Join(projectRoot, "test", "repos", name)
}

// GetTestConfigPath はテスト用設定ファイルのパスを返す。
func GetTestConfigPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filename))
	return filepath.Join(projectRoot, "test", "data", name)
}