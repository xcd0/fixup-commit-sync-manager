package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "FixupCommitSyncManager",
	Short: "A tool for synchronizing source files between Dev and Ops repositories",
	Long: `FixupCommitSyncManager is a comprehensive tool for Windows environments that provides:

- Automatic synchronization of source files between Dev and Ops repositories
- VHDX-based isolation and initialization features
- Configuration generation and validation
- Automated fixup commits with autosquash functionality
- Comprehensive logging and error handling

The tool supports various subcommands for different operations:
- init-config: Create configuration file through interactive wizard
- validate-config: Validate configuration file syntax and content
- init-vhdx: Initialize VHDX file with Ops repository
- mount-vhdx/unmount-vhdx: Mount/unmount VHDX files
- snapshot-vhdx: Manage VHDX snapshots
- sync: Synchronize files between repositories
- fixup: Perform fixup commits`,
	Version: "1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "Configuration file path (default: config.hjson)")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")

	rootCmd.AddCommand(NewInitConfigCmd())
	rootCmd.AddCommand(NewValidateConfigCmd())
	rootCmd.AddCommand(NewInitVHDXCmd())
	rootCmd.AddCommand(NewMountVHDXCmd())
	rootCmd.AddCommand(NewUnmountVHDXCmd())
	rootCmd.AddCommand(NewSnapshotVHDXCmd())
	rootCmd.AddCommand(NewSyncCmd())
	rootCmd.AddCommand(NewFixupCmd())
	rootCmd.AddCommand(NewHelpCmd())
}

func NewHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Display help information for subcommands",
		Long:  "Displays detailed help information for all available subcommands",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				rootCmd.Help()
				return
			}

			subCmd, _, err := rootCmd.Find(args)
			if err != nil {
				fmt.Printf("Unknown command: %s\n", args[0])
				rootCmd.Help()
				return
			}

			subCmd.Help()
		},
	}
}