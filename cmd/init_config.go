package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fixup-commit-sync-manager/internal/config"

	"github.com/spf13/cobra"
)

func NewInitConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-config",
		Short: "対話型ウィザードで HJSON 設定ファイルのテンプレートを作成",
		Long:  "対話型ウィザードを使用して必要な設定を収集し、新しい設定ファイルを作成します",
		RunE:  runInitConfig,
	}
}

func runInitConfig(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file %s already exists. Overwrite? (y/N): ", configPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Configuration file creation cancelled.")
			return nil
		}
	}

	cfg := gatherConfigInteractively()

	if err := writeConfigTemplate(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("Configuration file created successfully: %s\n", configPath)
	fmt.Println("Please review and modify the generated configuration as needed.")
	return nil
}

func gatherConfigInteractively() *config.Config {
	return gatherConfigInteractivelyWithWorkDir("", "")
}

func gatherConfigInteractivelyWithWorkDir(workDir, defaultLogPath string) *config.Config {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.DefaultConfig()

	fmt.Println("    === FixupCommitSyncManager 設定ウィザード ===")
	fmt.Println("    必要な設定を対話的に入力してください。")
	fmt.Println()

	// Devリポジトリパス（必須）
	fmt.Println("    【Devリポジトリ設定】")
	fmt.Println("    同期元となるDevリポジトリのローカルパスを指定してください。")
	fmt.Print("    Devリポジトリパス（必須）: ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.DevRepoPath = strings.TrimSpace(input)
	}

	// VHDXマウントポイント（必須）
	fmt.Println()
	fmt.Println("    【VHDX設定】")
	fmt.Println("    VHDXファイルをマウントするドライブレターを指定してください。")
	fmt.Println("    例: Q (Q:ドライブとしてマウント)")
	fmt.Print("    VHDXマウントポイント（必須） [X]: ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		mountPoint := strings.TrimSpace(input)
		if !strings.HasSuffix(mountPoint, ":") {
			mountPoint += ":"
		}
		cfg.MountPoint = mountPoint
	} else {
		cfg.MountPoint = "X:"
	}

	// Opsリポジトリパスを自動生成
	if cfg.DevRepoPath != "" {
		devBaseName := filepath.Base(cfg.DevRepoPath)
		cfg.OpsRepoPath = filepath.Join(cfg.MountPoint, devBaseName)
		fmt.Printf("    Opsリポジトリパス（自動生成）: %s\n", cfg.OpsRepoPath)
	}

	// VHDXサイズ
	fmt.Printf("    VHDXファイルサイズ [%s]: ", cfg.VHDXSize)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.VHDXSize = strings.TrimSpace(input)
	}

	// 同期間隔
	fmt.Println()
	fmt.Println("    【同期設定】")
	fmt.Println("    ファイル同期の実行間隔を指定してください。（例: 5m, 30s, 1h）")
	fmt.Printf("    同期間隔 [%s]: ", cfg.SyncInterval)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.SyncInterval = strings.TrimSpace(input)
	}

	// Fixup間隔
	fmt.Println("    Fixupコミットの実行間隔を指定してください。（例: 1h, 30m）")
	fmt.Printf("    Fixup間隔 [%s]: ", cfg.FixupInterval)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.FixupInterval = strings.TrimSpace(input)
	}

	// ログファイルパス
	fmt.Println()
	fmt.Println("    【ログ設定】")
	fmt.Println("    ログファイルの出力先を指定してください。")
	if defaultLogPath != "" {
		cfg.LogFilePath = defaultLogPath
	}
	fmt.Printf("    ログファイルパス [%s]: ", cfg.LogFilePath)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.LogFilePath = strings.TrimSpace(input)
	}

	// VHDX暗号化
	fmt.Println()
	fmt.Println("    【セキュリティ設定】")
	fmt.Println("    VHDXファイルの暗号化を有効にしますか？")
	fmt.Print("    VHDX暗号化を有効にする？ (y/N): ")
	if input, _ := reader.ReadString('\n'); strings.ToLower(strings.TrimSpace(input)) == "y" {
		cfg.EncryptionEnabled = true
	}

	// VHDXパスを自動生成
	if workDir != "" {
		cfg.VHDXPath = filepath.Join(workDir, "ops.vhdx")
	}

	return cfg
}

// runInitConfigImproved は改善された設定ファイル生成を実行する。
func runInitConfigImproved(configPath, workDir string) error {
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("    設定ファイル %s が既に存在します。上書きしますか？ (y/N): ", configPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("    設定ファイルの作成をキャンセルしました。")
			return nil
		}
	}

	// デフォルトのログファイルパスを生成
	defaultLogPath := "./sync.log"
	if workDir != "" {
		defaultLogPath = filepath.Join(workDir, "sync.log")
	}

	cfg := gatherConfigInteractivelyWithWorkDir(workDir, defaultLogPath)

	if err := writeConfigTemplate(configPath, cfg); err != nil {
		return fmt.Errorf("設定ファイルの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("\n    設定ファイルを作成しました: %s\n", configPath)
	fmt.Println("    必要に応じて設定ファイルを確認・編集してください。")
	return nil
}

func writeConfigTemplate(configPath string, cfg *config.Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := generateHJSONTemplate(cfg)
	return os.WriteFile(configPath, []byte(content), 0644)
}

func generateHJSONTemplate(cfg *config.Config) string {
	return fmt.Sprintf(`{
  // FixupCommitSyncManager Configuration File
  // This file uses HJSON format (Human JSON) which allows comments and relaxed syntax

  // === Repository Settings ===
  "devRepoPath": "%s",        // Dev repository local path (required)
  "opsRepoPath": "%s",        // Ops repository local path (required)

  // === File Synchronization Settings ===
  "includeExtensions": [".cpp", ".h", ".hpp"],  // File extensions to sync
  "includePatterns": [],      // Additional path patterns to include (Glob format)
  "excludePatterns": [],      // Path patterns to exclude (Glob format)

  // === Sync Operation Settings ===
  "syncInterval": "%s",       // Sync operation interval
  "pauseLockFile": "%s",      // Lock file name to pause sync
  "gitExecutable": "%s",      // Git command path
  "commitTemplate": "%s",     // Commit message template
  "authorName": "",           // Commit author name (empty = use git global config)
  "authorEmail": "",          // Commit author email (empty = use git global config)

  // === Fixup Settings ===
  "fixupInterval": "%s",      // Fixup commit interval
  "fixupMessagePrefix": "%s", // Fixup commit message prefix
  "autosquashEnabled": %t,    // Enable --autosquash flag
  // Note: Branch settings are now dynamic - automatically tracks Dev repository's current branch

  // === Retry and Error Handling ===
  "maxRetries": %d,           // Maximum retry attempts
  "retryDelay": "%s",         // Delay between retries
  "notifyOnError": {          // Error notification settings (optional)
    // "slackWebhookUrl": "https://hooks.slack.com/..."
  },

  // === Logging Settings ===
  "logLevel": "%s",           // Log level: DEBUG, INFO, WARN, ERROR
  "logFilePath": "%s",        // Log file output path
  "verbose": %t,              // Verbose output to stdout
  "dryRun": %t,               // Dry run mode (no actual operations)

  // === VHDX Settings ===
  "vhdxPath": "%s",           // VHDX file path (required for init-vhdx)
  "vhdxSize": "%s",           // VHDX file size
  "mountPoint": "%s",         // VHDX mount point (required for init-vhdx)
  "encryptionEnabled": %t     // Enable VHDX encryption
}`,
		cfg.DevRepoPath,
		cfg.OpsRepoPath,
		cfg.SyncInterval,
		cfg.PauseLockFile,
		cfg.GitExecutable,
		cfg.CommitTemplate,
		cfg.FixupInterval,
		cfg.FixupMsgPrefix,
		cfg.AutosquashEnabled,
		cfg.MaxRetries,
		cfg.RetryDelay,
		cfg.LogLevel,
		cfg.LogFilePath,
		cfg.Verbose,
		cfg.DryRun,
		cfg.VHDXPath,
		cfg.VHDXSize,
		cfg.MountPoint,
		cfg.EncryptionEnabled,
	)
}
