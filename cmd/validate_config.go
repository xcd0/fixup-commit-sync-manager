package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"fixup-commit-sync-manager/internal/config"

	"github.com/spf13/cobra"
)

func NewValidateConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-config",
		Short: "設定ファイルの構文と内容を検証",
		Long:  "HJSON 設定ファイルの構文エラー、必須フィールド、バージョン互換性を検証します",
		RunE:  runValidateConfig,
	}
}

func runValidateConfig(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Validating configuration file: %s\n", configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if verbose {
		fmt.Println("Configuration loaded successfully")
		fmt.Println("Validating configuration content...")
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := validatePaths(cfg, verbose); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	if err := validateVHDXConfig(cfg, verbose); err != nil {
		return fmt.Errorf("VHDX configuration validation failed: %w", err)
	}

	fmt.Println("✓ Configuration file is valid")
	
	if verbose {
		printConfigSummary(cfg)
	}

	return nil
}

func validatePaths(cfg *config.Config, verbose bool) error {
	if verbose {
		fmt.Println("Validating repository paths...")
	}

	if !filepath.IsAbs(cfg.DevRepoPath) {
		return fmt.Errorf("devRepoPath must be an absolute path: %s", cfg.DevRepoPath)
	}

	if !filepath.IsAbs(cfg.OpsRepoPath) {
		return fmt.Errorf("opsRepoPath must be an absolute path: %s", cfg.OpsRepoPath)
	}

	if cfg.DevRepoPath == cfg.OpsRepoPath {
		return fmt.Errorf("devRepoPath and opsRepoPath cannot be the same")
	}

	if verbose {
		if _, err := os.Stat(cfg.DevRepoPath); os.IsNotExist(err) {
			fmt.Printf("  Warning: Dev repository path does not exist: %s\n", cfg.DevRepoPath)
		} else {
			fmt.Printf("  ✓ Dev repository path exists: %s\n", cfg.DevRepoPath)
		}

		if _, err := os.Stat(cfg.OpsRepoPath); os.IsNotExist(err) {
			fmt.Printf("  Warning: Ops repository path does not exist: %s\n", cfg.OpsRepoPath)
		} else {
			fmt.Printf("  ✓ Ops repository path exists: %s\n", cfg.OpsRepoPath)
		}
	}

	return nil
}

func validateVHDXConfig(cfg *config.Config, verbose bool) error {
	if cfg.VHDXPath == "" {
		if verbose {
			fmt.Println("VHDX configuration not provided (optional)")
		}
		return nil
	}

	if verbose {
		fmt.Println("Validating VHDX configuration...")
	}

	if !filepath.IsAbs(cfg.VHDXPath) {
		return fmt.Errorf("vhdxPath must be an absolute path: %s", cfg.VHDXPath)
	}

	if cfg.MountPoint == "" {
		return fmt.Errorf("mountPoint is required when vhdxPath is specified")
	}

	vhdxDir := filepath.Dir(cfg.VHDXPath)
	if _, err := os.Stat(vhdxDir); os.IsNotExist(err) {
		return fmt.Errorf("VHDX directory does not exist: %s", vhdxDir)
	}

	if verbose {
		fmt.Printf("  ✓ VHDX path directory exists: %s\n", vhdxDir)
		fmt.Printf("  ✓ VHDX mount point configured: %s\n", cfg.MountPoint)
		fmt.Printf("  ✓ VHDX size configured: %s\n", cfg.VHDXSize)
		if cfg.EncryptionEnabled {
			fmt.Println("  ✓ VHDX encryption enabled")
		}
	}

	return nil
}

func printConfigSummary(cfg *config.Config) {
	fmt.Println("\n=== Configuration Summary ===")
	fmt.Printf("Dev Repository: %s\n", cfg.DevRepoPath)
	fmt.Printf("Ops Repository: %s\n", cfg.OpsRepoPath)
	fmt.Printf("Sync Interval: %s\n", cfg.SyncInterval)
	fmt.Printf("Fixup Interval: %s\n", cfg.FixupInterval)
	fmt.Printf("Target Branch: %s\n", cfg.TargetBranch)
	fmt.Printf("Base Branch: %s\n", cfg.BaseBranch)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log File: %s\n", cfg.LogFilePath)
	fmt.Printf("Include Extensions: %v\n", cfg.IncludeExtensions)
	
	if len(cfg.IncludePatterns) > 0 {
		fmt.Printf("Include Patterns: %v\n", cfg.IncludePatterns)
	}
	if len(cfg.ExcludePatterns) > 0 {
		fmt.Printf("Exclude Patterns: %v\n", cfg.ExcludePatterns)
	}
	
	if cfg.VHDXPath != "" {
		fmt.Printf("VHDX Path: %s\n", cfg.VHDXPath)
		fmt.Printf("VHDX Mount Point: %s\n", cfg.MountPoint)
		fmt.Printf("VHDX Size: %s\n", cfg.VHDXSize)
		fmt.Printf("VHDX Encryption: %t\n", cfg.EncryptionEnabled)
	}
	
	fmt.Printf("Dry Run: %t\n", cfg.DryRun)
	fmt.Printf("Verbose: %t\n", cfg.Verbose)
}