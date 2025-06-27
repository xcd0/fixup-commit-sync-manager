package cmd

import (
	"fmt"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/vhdx"

	"github.com/spf13/cobra"
)

func NewUnmountVHDXCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unmount-vhdx",
		Short: "VHDX ファイルをアンマウント",
		Long:  "指定された VHDX ファイルを設定されたマウントポイントからアンマウントします",
		RunE:  runUnmountVHDX,
	}
}

func runUnmountVHDX(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("vhdxPath is required in configuration for unmount-vhdx")
	}
	if cfg.MountPoint == "" {
		return fmt.Errorf("mountPoint is required in configuration for unmount-vhdx")
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Unmounting VHDX: %s\n", cfg.VHDXPath)
		fmt.Printf("Mount Point: %s\n", cfg.MountPoint)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would unmount VHDX file %s from %s\n", cfg.VHDXPath, cfg.MountPoint)
		return nil
	}

	if err := vhdxManager.UnmountVHDX(); err != nil {
		return fmt.Errorf("failed to unmount VHDX: %w", err)
	}

	fmt.Printf("✓ VHDX unmounted successfully from %s\n", cfg.MountPoint)
	return nil
}
