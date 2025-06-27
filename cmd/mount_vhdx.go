package cmd

import (
	"fmt"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/vhdx"

	"github.com/spf13/cobra"
)

func NewMountVHDXCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mount-vhdx",
		Short: "Mount VHDX file",
		Long:  "Mounts the specified VHDX file to the configured mount point",
		RunE:  runMountVHDX,
	}
}

func runMountVHDX(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose, _ := cmd.Flags().GetBool("verbose")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.VHDXPath == "" {
		return fmt.Errorf("vhdxPath is required in configuration for mount-vhdx")
	}
	if cfg.MountPoint == "" {
		return fmt.Errorf("mountPoint is required in configuration for mount-vhdx")
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Mounting VHDX: %s\n", cfg.VHDXPath)
		fmt.Printf("Mount Point: %s\n", cfg.MountPoint)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would mount VHDX file %s to %s\n", cfg.VHDXPath, cfg.MountPoint)
		return nil
	}

	if err := vhdxManager.MountVHDX(); err != nil {
		return fmt.Errorf("failed to mount VHDX: %w", err)
	}

	fmt.Printf("✓ VHDX mounted successfully at %s\n", cfg.MountPoint)
	return nil
}