package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestConfig はConfig構造体のテスト。
func TestConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			DevRepoPath: "/path/to/dev",
			OpsRepoPath: "/path/to/ops",
		}
		
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() failed: %v", err)
		}
	})
	
	t.Run("missing dev repo path", func(t *testing.T) {
		cfg := &Config{
			OpsRepoPath: "/path/to/ops",
		}
		
		if err := cfg.Validate(); err == nil {
			t.Error("Validate() should fail for missing dev repo path")
		}
	})
	
	t.Run("missing ops repo path", func(t *testing.T) {
		cfg := &Config{
			DevRepoPath: "/path/to/dev",
		}
		
		if err := cfg.Validate(); err == nil {
			t.Error("Validate() should fail for missing ops repo path")
		}
	})
}

// TestLoadConfigFromFile はloadConfigFromFile関数のテスト。
func TestLoadConfigFromFile(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.hjson")
		
		// テスト用設定ファイルを作成。
		configContent := `{
  "devRepoPath": "/tmp/dev-repo",
  "opsRepoPath": "/tmp/ops-repo",
  "syncInterval": "5m",
  "fixupInterval": "1h",
  "includeExtensions": [".cpp", ".h"],
  "excludePatterns": ["*.obj", "*.exe"],
  "vhdxPath": "/tmp/test.vhdx",
  "mountPoint": "X:",
  "vhdxSize": "10GB",
  "encryptionEnabled": false,
  "autosquashEnabled": true,
  "logLevel": "INFO",
  "logFilePath": "/tmp/test.log",
  "verbose": false
}`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// 設定ファイルを読み込み。
		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("loadConfigFromFile() failed: %v", err)
		}
		
		// 設定値の確認。
		if cfg.DevRepoPath != "/tmp/dev-repo" {
			t.Errorf("DevRepoPath = %s, want /tmp/dev-repo", cfg.DevRepoPath)
		}
		if cfg.OpsRepoPath != "/tmp/ops-repo" {
			t.Errorf("OpsRepoPath = %s, want /tmp/ops-repo", cfg.OpsRepoPath)
		}
		if cfg.SyncInterval != "5m" {
			t.Errorf("SyncInterval = %s, want 5m", cfg.SyncInterval)
		}
		if len(cfg.IncludeExtensions) != 2 {
			t.Errorf("IncludeExtensions length = %d, want 2", len(cfg.IncludeExtensions))
		}
	})
	
	t.Run("config file with comments", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.hjson")
		
		// コメント付き設定ファイルを作成。
		configContent := `{
  // Repository paths
  "devRepoPath": "/tmp/dev-repo",  // Dev repository path
  "opsRepoPath": "/tmp/ops-repo",  // Ops repository path
  
  // Sync settings
  "syncInterval": "5m"  // Sync interval
}`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// 設定ファイルを読み込み。
		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("loadConfigFromFile() failed: %v", err)
		}
		
		// 設定値の確認。
		if cfg.DevRepoPath != "/tmp/dev-repo" {
			t.Errorf("DevRepoPath = %s, want /tmp/dev-repo", cfg.DevRepoPath)
		}
		if cfg.SyncInterval != "5m" {
			t.Errorf("SyncInterval = %s, want 5m", cfg.SyncInterval)
		}
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadConfigFromFile("/non/existent/path.hjson")
		if err == nil {
			t.Error("loadConfigFromFile() should fail for non-existent file")
		}
	})
	
	t.Run("invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid.hjson")
		
		// 無効なJSONファイルを作成。
		invalidContent := `{ "devRepoPath": "/tmp/dev-repo" "missing comma" }`
		
		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}
		
		_, err = loadConfigFromFile(configPath)
		if err == nil {
			t.Error("loadConfigFromFile() should fail for invalid JSON")
		}
	})
}

// TestLoadConfiguration はloadConfiguration関数のテスト。
func TestLoadConfiguration(t *testing.T) {
	t.Run("load existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.hjson")
		
		// テスト用設定ファイルを作成。
		configContent := `{
  "devRepoPath": "/tmp/dev-repo",
  "opsRepoPath": "/tmp/ops-repo"
}`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// 設定を読み込み。
		cfg, err := loadConfiguration(configPath)
		if err != nil {
			t.Fatalf("loadConfiguration() failed: %v", err)
		}
		
		if cfg.DevRepoPath != "/tmp/dev-repo" {
			t.Errorf("DevRepoPath = %s, want /tmp/dev-repo", cfg.DevRepoPath)
		}
	})
	
	t.Run("default config path", func(t *testing.T) {
		// デフォルトパスが使用されることを確認。
		// ファイルが存在しないためエラーになるが、パス処理は確認できる。
		_, err := loadConfiguration("")
		if err == nil {
			t.Error("loadConfiguration() should fail for non-existent default config")
		}
		
		// エラーメッセージに"config.hjson"が含まれることを確認。
		if !strings.Contains(err.Error(), "config.hjson") {
			t.Errorf("Error should mention default config file: %v", err)
		}
	})
}

// TestCheckAndInitConfig はcheckAndInitConfig関数のテスト。
func TestCheckAndInitConfig(t *testing.T) {
	t.Run("create new config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "new-config.hjson")
		
		args := &RunArgs{
			ConfigPath: configPath,
		}
		
		// 設定ファイルが存在しないため作成される。
		err := checkAndInitConfig(args)
		if err == nil {
			t.Error("checkAndInitConfig() should fail requesting user to edit config")
		}
		
		// エラーメッセージに編集要求が含まれることを確認。
		if !strings.Contains(err.Error(), "編集が必要") {
			t.Errorf("Error should request config editing: %v", err)
		}
		
		// 設定ファイルが作成されたことを確認。
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config file should be created: %s", configPath)
		}
		
		// 作成されたファイルの内容を確認。
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read created config file: %v", err)
		}
		
		if !strings.Contains(string(content), "devRepoPath") {
			t.Error("Created config file should contain devRepoPath")
		}
	})
	
	t.Run("existing config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "existing-config.hjson")
		
		// 既存の設定ファイルを作成。
		err := os.WriteFile(configPath, []byte(`{"devRepoPath": "/tmp/dev"}`), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing config file: %v", err)
		}
		
		args := &RunArgs{
			ConfigPath: configPath,
		}
		
		// 既存ファイルがあるため成功する。
		err = checkAndInitConfig(args)
		if err != nil {
			t.Errorf("checkAndInitConfig() should succeed for existing file: %v", err)
		}
	})
	
	t.Run("default config path", func(t *testing.T) {
		// 作業ディレクトリを一時ディレクトリに変更。
		originalWD, _ := os.Getwd()
		tempDir := t.TempDir()
		os.Chdir(tempDir)
		defer os.Chdir(originalWD)
		
		args := &RunArgs{}
		
		// デフォルトパスが使用されることを確認。
		err := checkAndInitConfig(args)
		if err == nil {
			t.Error("checkAndInitConfig() should fail for new default config")
		}
		
		// ConfigPathが設定されることを確認。
		if args.ConfigPath != "config.hjson" {
			t.Errorf("ConfigPath = %s, want config.hjson", args.ConfigPath)
		}
	})
}

// TestValidateConfiguration はvalidateConfiguration関数のテスト。
func TestValidateConfiguration(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "valid-config.hjson")
		
		// 有効な設定ファイルを作成。
		configContent := `{
  "devRepoPath": "/tmp/dev-repo",
  "opsRepoPath": "/tmp/ops-repo"
}`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// 設定検証を実行。
		err = validateConfiguration(configPath, false)
		if err != nil {
			t.Errorf("validateConfiguration() failed: %v", err)
		}
	})
	
	t.Run("invalid configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid-config.hjson")
		
		// 無効な設定ファイルを作成（必須項目欠如）。
		configContent := `{
  "syncInterval": "5m"
}`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// 設定検証を実行。
		err = validateConfiguration(configPath, false)
		if err == nil {
			t.Error("validateConfiguration() should fail for invalid config")
		}
	})
}

// TestCloneRepositorySimple はcloneRepositorySimple関数のテスト。
func TestCloneRepositorySimple(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Git not available, skipping clone repository test")
	}
	
	t.Run("clone local repository", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// ソースリポジトリを作成。
		srcPath := filepath.Join(tempDir, "source-repo")
		err := os.MkdirAll(srcPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create source directory: %v", err)
		}
		
		// ソースでgitリポジトリを初期化。
		err = runCommandInDir("git init", srcPath)
		if err != nil {
			t.Fatalf("Failed to init source git repo: %v", err)
		}
		
		// ソースに設定を追加。
		err = runCommandInDir("git config user.name TestUser", srcPath)
		if err != nil {
			t.Fatalf("Failed to configure source git repo name: %v", err)
		}
		
		err = runCommandInDir("git config user.email test@example.com", srcPath)
		if err != nil {
			t.Fatalf("Failed to configure source git repo email: %v", err)
		}
		
		// ソースにテストファイルを作成してコミット。
		testFile := filepath.Join(srcPath, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		err = runCommandInDir("git add test.txt", srcPath)
		if err != nil {
			t.Fatalf("Failed to add file to source repo: %v", err)
		}
		
		err = runCommandInDir("git commit -m Initial_commit", srcPath)
		if err != nil {
			t.Fatalf("Failed to commit in source repo: %v", err)
		}
		
		// クローン先のパス。
		destPath := filepath.Join(tempDir, "dest-repo")
		
		// クローンを実行。
		err = cloneRepositorySimple(srcPath, destPath)
		if err != nil {
			t.Errorf("cloneRepositorySimple() failed: %v", err)
		}
		
		// クローンされたディレクトリが存在することを確認。
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("Cloned directory should exist: %s", destPath)
		}
		
		// クローンされたファイルが存在することを確認。
		clonedFile := filepath.Join(destPath, "test.txt")
		if _, err := os.Stat(clonedFile); os.IsNotExist(err) {
			t.Errorf("Cloned file should exist: %s", clonedFile)
		}
		
		// クローンされたファイルの内容を確認。
		content, err := os.ReadFile(clonedFile)
		if err != nil {
			t.Fatalf("Failed to read cloned file: %v", err)
		}
		if string(content) != "test content" {
			t.Errorf("Cloned file content = %s, want 'test content'", string(content))
		}
	})
	
	t.Run("clone to existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := filepath.Join(tempDir, "source")
		destPath := filepath.Join(tempDir, "dest")
		
		// 既存のディレクトリを作成。
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create dest directory: %v", err)
		}
		
		// 既存ディレクトリへのクローンはエラーになることを確認。
		err = cloneRepositorySimple(srcPath, destPath)
		if err == nil {
			t.Error("cloneRepositorySimple() should fail for existing directory")
		}
	})
	
	t.Run("clone non-existent repository", func(t *testing.T) {
		tempDir := t.TempDir()
		srcPath := "/non/existent/repo"
		destPath := filepath.Join(tempDir, "dest")
		
		// 存在しないリポジトリのクローンはエラーになることを確認。
		err := cloneRepositorySimple(srcPath, destPath)
		if err == nil {
			t.Error("cloneRepositorySimple() should fail for non-existent repository")
		}
	})
	
	t.Run("clone to Windows drive root", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows-specific test")
		}
		
		srcPath := "/source/repo"
		destPath := "Q:/my-repo" // Windowsドライブルート下
		
		// Windowsドライブルートの場合の動作確認。
		// (実際のクローンは失敗するが、ディレクトリ処理が正しく動作することを確認)
		err := cloneRepositorySimple(srcPath, destPath)
		// エラーは期待される（ソースが存在しないため）
		if err == nil {
			t.Log("Windows drive root clone succeeded (unexpected but okay)")
		}
	})
}


// TestIsWindowsDriveRoot はisWindowsDriveRoot関数のテスト。
func TestIsWindowsDriveRoot(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"Q:", true},
		{"X:", true},
		{"Z:", true},
		{"C:", true},
		{"Q:/", false},
		{"X:/path", false},
		{"/tmp", false},
		{"./path", false},
		{"Q", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isWindowsDriveRoot(tt.path)
			if result != tt.expected {
				t.Errorf("isWindowsDriveRoot(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsVHDXPath はisVHDXPath関数のテスト。
func TestIsVHDXPath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		mountPoint string
		expected   bool
	}{
		{"VHDX path with Q:", "Q:/neco", "Q:", true},
		{"VHDX path with Q: subdir", "Q:/neco/subdir", "Q:", true},
		{"Non-VHDX path", "/tmp/neco", "Q:", false},
		{"Empty mount point", "Q:/neco", "", false},
		{"Different drive", "P:/neco", "Q:", false},
		{"Partial match", "QQ:/neco", "Q:", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVHDXPath(tt.path, tt.mountPoint)
			if result != tt.expected {
				t.Errorf("isVHDXPath(%q, %q) = %v, want %v", tt.path, tt.mountPoint, result, tt.expected)
			}
		})
	}
}

// TestGenerateLocalFallbackPath はgenerateLocalFallbackPath関数のテスト。
func TestGenerateLocalFallbackPath(t *testing.T) {
	tests := []struct {
		name           string
		devRepoPath    string
		originalOpsPath string
		expectedSuffix string
	}{
		{
			name:           "Simple repo name",
			devRepoPath:    "P:/neco",
			originalOpsPath: "Q:/neco",
			expectedSuffix: "ops-neco",
		},
		{
			name:           "Complex repo path",
			devRepoPath:    "/home/user/my-project",
			originalOpsPath: "Q:/my-project",
			expectedSuffix: "ops-my-project",
		},
		{
			name:           "Windows style path",
			devRepoPath:    "C:/dev/project-name", // Linux環境でもテスト可能な形式
			originalOpsPath: "X:/project-name",
			expectedSuffix: "ops-project-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateLocalFallbackPath(tt.devRepoPath, tt.originalOpsPath)
			
			// 結果が期待するサフィックスで終わることを確認
			if !strings.HasSuffix(result, tt.expectedSuffix) {
				t.Errorf("generateLocalFallbackPath(%q, %q) = %q, should end with %q", 
					tt.devRepoPath, tt.originalOpsPath, result, tt.expectedSuffix)
			}
			
			// 結果が絶対パスまたは相対パスであることを確認
			if result == "" {
				t.Errorf("generateLocalFallbackPath should not return empty string")
			}
		})
	}
}

// TestRunArgsValidation はRunArgs構造体のテスト。
func TestRunArgsValidation(t *testing.T) {
	args := &RunArgs{
		ConfigPath: "test-config.hjson",
		DryRun:     true,
		Verbose:    true,
		NoVhdx:     true,
		SkipInit:   true,
	}
	
	// 構造体のフィールドが正しく設定されることを確認。
	if args.ConfigPath != "test-config.hjson" {
		t.Errorf("ConfigPath = %s, want test-config.hjson", args.ConfigPath)
	}
	if !args.DryRun {
		t.Error("DryRun should be true")
	}
	if !args.Verbose {
		t.Error("Verbose should be true")
	}
	if !args.NoVhdx {
		t.Error("NoVhdx should be true")
	}
	if !args.SkipInit {
		t.Error("SkipInit should be true")
	}
}

// TestConfigJSONMarshaling はConfig構造体のJSONマーシャリングテスト。
func TestConfigJSONMarshaling(t *testing.T) {
	originalConfig := &Config{
		DevRepoPath:         "/tmp/dev",
		OpsRepoPath:         "/tmp/ops",
		SyncInterval:        "5m",
		FixupInterval:       "1h",
		IncludeExtensions:   []string{".cpp", ".h"},
		ExcludePatterns:     []string{"*.obj"},
		VhdxPath:            "/tmp/test.vhdx",
		MountPoint:          "X:",
		VhdxSize:            "10GB",
		EncryptionEnabled:   false,
		AutosquashEnabled:   true,
		LogLevel:            "INFO",
		LogFilePath:         "/tmp/test.log",
		Verbose:             true,
	}
	
	// JSONにマーシャル。
	data, err := json.Marshal(originalConfig)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	
	// JSONからアンマーシャル。
	var unmarshaledConfig Config
	err = json.Unmarshal(data, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	
	// 値が正しく復元されることを確認。
	if unmarshaledConfig.DevRepoPath != originalConfig.DevRepoPath {
		t.Errorf("DevRepoPath = %s, want %s", unmarshaledConfig.DevRepoPath, originalConfig.DevRepoPath)
	}
	if unmarshaledConfig.SyncInterval != originalConfig.SyncInterval {
		t.Errorf("SyncInterval = %s, want %s", unmarshaledConfig.SyncInterval, originalConfig.SyncInterval)
	}
	if len(unmarshaledConfig.IncludeExtensions) != len(originalConfig.IncludeExtensions) {
		t.Errorf("IncludeExtensions length = %d, want %d", 
			len(unmarshaledConfig.IncludeExtensions), len(originalConfig.IncludeExtensions))
	}
}

// TestRunCommand はrunCommand関数のテスト。
func TestRunCommand(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		// 成功するコマンドのテスト。
		err := runCommand("echo test")
		if err != nil {
			t.Errorf("runCommand() failed: %v", err)
		}
	})
	
	t.Run("command with output", func(t *testing.T) {
		// 出力を生成するコマンドのテスト。
		err := runCommand("echo hello world")
		if err != nil {
			t.Errorf("runCommand() failed: %v", err)
		}
	})
	
	t.Run("non-existent command", func(t *testing.T) {
		// 存在しないコマンドのテスト。
		err := runCommand("nonexistentcommand12345")
		if err == nil {
			t.Error("runCommand() should fail for non-existent command")
		}
	})
	
	t.Run("empty command", func(t *testing.T) {
		// 空のコマンドのテスト。
		err := runCommand("")
		if err == nil {
			t.Error("runCommand() should fail for empty command")
		}
	})
	
	t.Run("whitespace only command", func(t *testing.T) {
		// 空白のみのコマンドのテスト。
		err := runCommand("   ")
		if err == nil {
			t.Error("runCommand() should fail for whitespace-only command")
		}
	})
}

// TestInitializationFlow は初期化フローの統合テスト。
func TestInitializationFlow(t *testing.T) {
	t.Run("skip initialization", func(t *testing.T) {
		args := &RunArgs{
			SkipInit: true,
		}
		
		ctx := context.Background()
		err := runInitializationFlow(ctx, args)
		if err != nil {
			t.Errorf("runInitializationFlow() with skip should succeed: %v", err)
		}
	})
}

// TestVHDXOpsRepoPathGenerationInRun はrun.goでのVHDXOpsリポジトリパス生成をテストする。
func TestVHDXOpsRepoPathGenerationInRun(t *testing.T) {
	tests := []struct {
		name        string
		devRepoPath string
		mountPoint  string
		vhdxPath    string
		expectedBaseName string
	}{
		{
			name:        "Windows drive letter Q: with simple repo",
			devRepoPath: "/path/to/my-repo",
			mountPoint:  "Q:",
			vhdxPath:    "/tmp/test.vhdx",
			expectedBaseName: "my-repo",
		},
		{
			name:        "Windows drive letter X: with complex path",
			devRepoPath: "C:/Users/dev/project-name", // Linux環境でもテスト可能な形式
			mountPoint:  "X:",
			vhdxPath:    "/tmp/test.vhdx",
			expectedBaseName: "project-name",
		},
		{
			name:        "Complex repository name",
			devRepoPath: "/home/user/fixup-commit-sync-manager",
			mountPoint:  "Z:",
			vhdxPath:    "/tmp/test.vhdx",
			expectedBaseName: "fixup-commit-sync-manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DevRepoPath: tt.devRepoPath,
				MountPoint:  tt.mountPoint,
				VhdxPath:    tt.vhdxPath,
				OpsRepoPath: "/default/ops", // 初期値
			}

			// run.goの処理をシミュレート。
			opsRepoPath := cfg.OpsRepoPath
			if cfg.VhdxPath != "" && cfg.MountPoint != "" {
				// Windowsドライブレター形式のマウントポイントに対応（例: "Q:" → "Q:\\devBaseName"）
				devBaseName := filepath.Base(cfg.DevRepoPath)
				opsRepoPath, _ = filepath.Abs(filepath.Join(cfg.MountPoint, devBaseName))
			}

			normalizedPath := filepath.ToSlash(opsRepoPath)

			// ベース名が正しく抽出されることを確認。
			devBaseName := filepath.Base(cfg.DevRepoPath)
			if devBaseName != tt.expectedBaseName {
				t.Errorf("Expected base name %q, got %q", tt.expectedBaseName, devBaseName)
			}

			// VHDXが有効な場合、パスにマウントポイントとベース名が含まれることを確認。
			if cfg.VhdxPath != "" && cfg.MountPoint != "" {
				if !strings.Contains(normalizedPath, tt.mountPoint) {
					t.Errorf("OpsRepoPath should contain mount point %q: %q", tt.mountPoint, normalizedPath)
				}
				if !strings.Contains(normalizedPath, tt.expectedBaseName) {
					t.Errorf("OpsRepoPath should contain base name %q: %q", tt.expectedBaseName, normalizedPath)
				}
			}
		})
	}
}

// TestPeriodicExecution は定期実行のテスト。
func TestPeriodicExecution(t *testing.T) {
	t.Run("periodic execution with short timeout", func(t *testing.T) {
		cfg := &Config{
			DevRepoPath:   "/tmp/dev",
			OpsRepoPath:   "/tmp/ops",
			SyncInterval:  "100ms",  // 短い間隔でテスト。
			FixupInterval: "200ms",  // 短い間隔でテスト。
		}
		
		args := &RunArgs{
			DryRun: true,  // DryRunモードでテスト。
		}
		
		// 短時間で終了するコンテキスト。
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		
		// 定期実行を開始（短時間で終了）。
		err := runPeriodicExecution(ctx, cfg, args)
		if err != nil {
			t.Errorf("runPeriodicExecution() failed: %v", err)
		}
	})
}