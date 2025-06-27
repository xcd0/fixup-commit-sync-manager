package cmd

import (
	"fmt"
	"time"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/sync"

	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize source files between Dev and Ops repositories",
		Long:  "Synchronizes tracked and new source files from Dev repository to Ops repository and commits changes",
		RunE:  runSync,
	}

	cmd.Flags().Bool("continuous", false, "Run sync continuously at configured interval")
	
	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
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

	syncer := sync.NewFileSyncer(cfg)

	if continuous {
		return runContinuousSync(syncer, cfg)
	}

	return runSingleSync(syncer, cfg)
}

func runSingleSync(syncer *sync.FileSyncer, cfg *config.Config) error {
	if cfg.Verbose {
		fmt.Println("Starting single sync operation...")
		fmt.Printf("Dev Repository: %s\n", cfg.DevRepoPath)
		fmt.Printf("Ops Repository: %s\n", cfg.OpsRepoPath)
	}

	if cfg.DryRun {
		fmt.Println("[DRY RUN] Would perform sync operation")
		return nil
	}

	result, err := syncer.Sync()
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if len(result.FilesAdded)+len(result.FilesModified)+len(result.FilesDeleted) == 0 {
		if cfg.Verbose {
			fmt.Println("No changes detected - sync skipped")
		}
		return nil
	}

	fmt.Printf("✓ Sync completed successfully\n")
	fmt.Printf("  Files added: %d\n", len(result.FilesAdded))
	fmt.Printf("  Files modified: %d\n", len(result.FilesModified))
	fmt.Printf("  Files deleted: %d\n", len(result.FilesDeleted))
	
	if result.CommitHash != "" {
		fmt.Printf("  Commit: %s\n", result.CommitHash[:8])
	}

	if cfg.Verbose && len(result.FilesAdded) > 0 {
		fmt.Println("Added files:")
		for _, file := range result.FilesAdded {
			fmt.Printf("  + %s\n", file)
		}
	}

	if cfg.Verbose && len(result.FilesModified) > 0 {
		fmt.Println("Modified files:")
		for _, file := range result.FilesModified {
			fmt.Printf("  ~ %s\n", file)
		}
	}

	if cfg.Verbose && len(result.FilesDeleted) > 0 {
		fmt.Println("Deleted files:")
		for _, file := range result.FilesDeleted {
			fmt.Printf("  - %s\n", file)
		}
	}

	return nil
}

func runContinuousSync(syncer *sync.FileSyncer, cfg *config.Config) error {
	interval, err := cfg.GetSyncIntervalDuration()
	if err != nil {
		return fmt.Errorf("invalid sync interval: %w", err)
	}

	fmt.Printf("Starting continuous sync with interval: %s\n", cfg.SyncInterval)
	fmt.Printf("Dev Repository: %s\n", cfg.DevRepoPath)
	fmt.Printf("Ops Repository: %s\n", cfg.OpsRepoPath)
	fmt.Println("Press Ctrl+C to stop")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if cfg.Verbose {
				fmt.Printf("\n[%s] Starting sync operation...\n", time.Now().Format("15:04:05"))
			}

			if cfg.DryRun {
				fmt.Printf("[%s] [DRY RUN] Would perform sync operation\n", time.Now().Format("15:04:05"))
				continue
			}

			result, err := syncer.Sync()
			if err != nil {
				fmt.Printf("[%s] Sync failed: %v\n", time.Now().Format("15:04:05"), err)
				continue
			}

			if len(result.FilesAdded)+len(result.FilesModified)+len(result.FilesDeleted) == 0 {
				if cfg.Verbose {
					fmt.Printf("[%s] No changes detected\n", time.Now().Format("15:04:05"))
				}
				continue
			}

			fmt.Printf("[%s] ✓ Sync completed - Files: +%d ~%d -%d", 
				time.Now().Format("15:04:05"),
				len(result.FilesAdded), 
				len(result.FilesModified), 
				len(result.FilesDeleted))
			
			if result.CommitHash != "" {
				fmt.Printf(" Commit: %s", result.CommitHash[:8])
			}
			fmt.Println()
		}
	}
}