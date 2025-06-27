package cmd

import (
	"fmt"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/fixup"

	"github.com/spf13/cobra"
)

func NewFixupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fixup",
		Short: "Perform fixup commit operations on Ops repository",
		Long:  "Creates fixup commits for changes in the Ops repository and optionally performs autosquash rebase",
		RunE:  runFixup,
	}

	cmd.Flags().Bool("continuous", false, "Run fixup continuously at configured interval")
	
	return cmd
}

func runFixup(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose, _ := cmd.Flags().GetBool("verbose")
	continuous, _ := cmd.Flags().GetBool("continuous")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if dryRun {
		cfg.DryRun = true
	}
	if verbose {
		cfg.Verbose = true
	}

	fixupManager := fixup.NewFixupManager(cfg)

	if continuous {
		return fixupManager.RunContinuousFixup()
	}

	return runSingleFixup(fixupManager, cfg)
}

func runSingleFixup(fixupManager *fixup.FixupManager, cfg *config.Config) error {
	if cfg.Verbose {
		fmt.Println("Starting fixup operation...")
		fmt.Printf("Ops Repository: %s\n", cfg.OpsRepoPath)
		fmt.Printf("Target Branch: %s\n", cfg.TargetBranch)
		fmt.Printf("Base Branch: %s\n", cfg.BaseBranch)
		fmt.Printf("Autosquash Enabled: %t\n", cfg.AutosquashEnabled)
	}

	if cfg.DryRun {
		fmt.Println("[DRY RUN] Would perform fixup operation")
		return nil
	}

	result, err := fixupManager.RunFixup()
	if err != nil {
		return fmt.Errorf("fixup failed: %w", err)
	}

	if result.FilesModified == 0 {
		if cfg.Verbose {
			fmt.Println("No uncommitted changes found - fixup skipped")
		} else {
			fmt.Println("No changes to fixup")
		}
		return nil
	}

	fmt.Printf("✓ Fixup completed successfully\n")
	fmt.Printf("  Files modified: %d\n", result.FilesModified)
	
	if result.FixupCommitHash != "" {
		fmt.Printf("  Fixup commit: %s\n", result.FixupCommitHash[:8])
	}
	
	if result.CommitHash != "" {
		fmt.Printf("  Base commit: %s\n", result.CommitHash[:8])
	}

	if cfg.AutosquashEnabled {
		fmt.Println("  Autosquash rebase: completed")
	}

	if cfg.Verbose {
		fmt.Printf("Fixup message prefix: %s\n", cfg.FixupMsgPrefix)
	}

	return nil
}