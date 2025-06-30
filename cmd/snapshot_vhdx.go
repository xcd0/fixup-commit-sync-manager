package cmd

import (
	"fmt"
	"strings"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/vhdx"

	"github.com/spf13/cobra"
)

func NewSnapshotVHDXCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot-vhdx",
		Short: "VHDX スナップショットを管理",
		Long:  "VHDX スナップショットの作成、一覧表示、ロールバックを実行します",
	}

	cmd.AddCommand(NewCreateSnapshotCmd())
	cmd.AddCommand(NewListSnapshotsCmd())
	cmd.AddCommand(NewRollbackSnapshotCmd())

	return cmd
}

func NewCreateSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [snapshot-name]",
		Short: "新しい VHDX スナップショットを作成",
		Long:  "指定された名前で VHDX ファイルの新しいスナップショットを作成します",
		RunE:  runCreateSnapshot,
	}
	return cmd
}

func NewListSnapshotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "利用可能な VHDX スナップショットを一覧表示",
		Long:  "VHDX ファイルの利用可能なすべてのスナップショットを一覧表示します",
		RunE:  runListSnapshots,
	}
	return cmd
}

func NewRollbackSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback <snapshot-name>",
		Short: "VHDX スナップショットにロールバック",
		Long:  "VHDX ファイルを指定されたスナップショットにロールバックします",
		Args:  cobra.ExactArgs(1),
		RunE:  runRollbackSnapshot,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return getAvailableSnapshots(cmd), cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}

func runCreateSnapshot(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("vhdxPath is required in configuration for snapshot operations")
	}

	var snapshotName string
	if len(args) > 0 {
		snapshotName = args[0]
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Creating snapshot of VHDX: %s\n", cfg.VHDXPath)
		if snapshotName != "" {
			fmt.Printf("Snapshot name: %s\n", snapshotName)
		}
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would create snapshot of %s", cfg.VHDXPath)
		if snapshotName != "" {
			fmt.Printf(" with name %s", snapshotName)
		}
		fmt.Println()
		return nil
	}

	if err := vhdxManager.CreateSnapshot(snapshotName); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	if snapshotName != "" {
		fmt.Printf("✓ Snapshot '%s' created successfully\n", snapshotName)
	} else {
		fmt.Println("✓ Snapshot created successfully")
	}
	return nil
}

func runListSnapshots(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.VHDXPath == "" {
		return fmt.Errorf("vhdxPath is required in configuration for snapshot operations")
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Listing snapshots for VHDX: %s\n", cfg.VHDXPath)
	}

	snapshots, err := vhdxManager.ListSnapshots()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found")
		return nil
	}

	fmt.Printf("Available snapshots (%d):\n", len(snapshots))
	for i, snapshot := range snapshots {
		fmt.Printf("  %d. %s\n", i+1, snapshot)
	}

	return nil
}

// getAvailableSnapshots は利用可能なスナップショット名を取得する。
func getAvailableSnapshots(cmd *cobra.Command) []string {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "config.hjson"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return []string{}
	}

	if cfg.VHDXPath == "" {
		return []string{}
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)
	snapshots, err := vhdxManager.ListSnapshots()
	if err != nil {
		return []string{}
	}

	return snapshots
}

func runRollbackSnapshot(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("vhdxPath is required in configuration for snapshot operations")
	}

	snapshotName := args[0]
	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Rolling back VHDX: %s\n", cfg.VHDXPath)
		fmt.Printf("Target snapshot: %s\n", snapshotName)
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would rollback %s to snapshot '%s'\n", cfg.VHDXPath, snapshotName)
		return nil
	}

	fmt.Printf("⚠️  WARNING: This will replace the current VHDX with snapshot '%s'\n", snapshotName)
	fmt.Print("Are you sure you want to continue? (yes/no): ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "yes" && strings.ToLower(response) != "y" {
		fmt.Println("Rollback cancelled")
		return nil
	}

	if err := vhdxManager.RollbackToSnapshot(snapshotName); err != nil {
		return fmt.Errorf("failed to rollback to snapshot: %w", err)
	}

	fmt.Printf("✓ Successfully rolled back to snapshot '%s'\n", snapshotName)
	return nil
}
