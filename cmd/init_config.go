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
	reader := bufio.NewReader(os.Stdin)
	cfg := config.DefaultConfig()

	fmt.Println("=== FixupCommitSyncManager Configuration Wizard ===")
	fmt.Println()

	fmt.Print("Enter Dev repository path (required): ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.DevRepoPath = strings.TrimSpace(input)
	}

	fmt.Print("Enter Ops repository path (required): ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.OpsRepoPath = strings.TrimSpace(input)
	}

	fmt.Printf("Enter sync interval [%s]: ", cfg.SyncInterval)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.SyncInterval = strings.TrimSpace(input)
	}

	fmt.Printf("Enter fixup interval [%s]: ", cfg.FixupInterval)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.FixupInterval = strings.TrimSpace(input)
	}

	// ブランチ設定は動的追従により不要になった。
	fmt.Println("Branch configuration: Using dynamic tracking from Dev repository (no manual configuration needed)")

	fmt.Printf("Enter log file path [%s]: ", cfg.LogFilePath)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.LogFilePath = strings.TrimSpace(input)
	}

	fmt.Print("Enter VHDX file path (optional): ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.VHDXPath = strings.TrimSpace(input)
	}

	fmt.Print("Enter VHDX mount point (optional, e.g., X:): ")
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.MountPoint = strings.TrimSpace(input)
	}

	fmt.Printf("Enter VHDX size [%s]: ", cfg.VHDXSize)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		cfg.VHDXSize = strings.TrimSpace(input)
	}

	fmt.Print("Enable VHDX encryption? (y/N): ")
	if input, _ := reader.ReadString('\n'); strings.ToLower(strings.TrimSpace(input)) == "y" {
		cfg.EncryptionEnabled = true
	}

	return cfg
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
