package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fixup-commit-sync-manager/internal/config"
	"fixup-commit-sync-manager/internal/vhdx"

	"github.com/spf13/cobra"
)

func NewInitVHDXCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-vhdx",
		Short: "Ops リポジトリで VHDX ファイルを初期化",
		Long:  "VHDX ファイルを作成し、マウント、リポジトリクローン、リモート URL 設定、アンマウントを実行します",
		RunE:  runInitVHDX,
	}
}

func runInitVHDX(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("vhdxPath is required in configuration for init-vhdx")
	}
	if cfg.MountPoint == "" {
		return fmt.Errorf("mountPoint is required in configuration for init-vhdx")
	}
	if cfg.DevRepoPath == "" {
		return fmt.Errorf("devRepoPath is required in configuration for init-vhdx")
	}

	vhdxManager := vhdx.NewVHDXManager(cfg.VHDXPath, cfg.MountPoint, cfg.VHDXSize, cfg.EncryptionEnabled)

	if verbose {
		fmt.Printf("Initializing VHDX: %s\n", cfg.VHDXPath)
		fmt.Printf("Mount Point: %s\n", cfg.MountPoint)
		fmt.Printf("Size: %s\n", cfg.VHDXSize)
		fmt.Printf("Encryption: %t\n", cfg.EncryptionEnabled)
	}

	if dryRun {
		fmt.Println("[DRY RUN] Would perform the following operations:")
		fmt.Printf("  1. Create VHDX file: %s\n", cfg.VHDXPath)
		fmt.Printf("  2. Mount VHDX to: %s\n", cfg.MountPoint)
		fmt.Printf("  3. Clone repository from: %s\n", cfg.DevRepoPath)
		fmt.Printf("  4. Set up Ops repository at: %s\n", filepath.Join(cfg.MountPoint, "ops-repo"))
		fmt.Printf("  5. Unmount VHDX\n")
		return nil
	}

	if verbose {
		fmt.Println("Step 1: Creating VHDX file...")
	}
	if err := vhdxManager.CreateVHDX(); err != nil {
		return fmt.Errorf("failed to create VHDX: %w", err)
	}

	if verbose {
		fmt.Println("Step 2: Mounting VHDX...")
	}
	if err := vhdxManager.MountVHDX(); err != nil {
		return fmt.Errorf("failed to mount VHDX: %w", err)
	}

	defer func() {
		if verbose {
			fmt.Println("Step 5: Unmounting VHDX...")
		}
		if unmountErr := vhdxManager.UnmountVHDX(); unmountErr != nil {
			fmt.Printf("Warning: Failed to unmount VHDX: %v\n", unmountErr)
		}
	}()

	opsRepoPath := filepath.Join(cfg.MountPoint, "ops-repo")
	
	if verbose {
		fmt.Printf("Step 3: Cloning repository to: %s\n", opsRepoPath)
	}
	if err := cloneRepository(cfg.DevRepoPath, opsRepoPath, cfg.GitExecutable); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	if verbose {
		fmt.Println("Step 4: Setting up remote URL and branch...")
	}
	if err := setupOpsRepository(opsRepoPath, cfg); err != nil {
		return fmt.Errorf("failed to setup Ops repository: %w", err)
	}

	fmt.Printf("✓ VHDX initialization completed successfully\n")
	fmt.Printf("  VHDX File: %s\n", cfg.VHDXPath)
	fmt.Printf("  Ops Repository will be available at: %s when mounted\n", opsRepoPath)
	
	return nil
}

func cloneRepository(sourcePath, targetPath string, gitExecutable string) error {
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source repository does not exist: %s", sourcePath)
	}

	gitDir := filepath.Join(sourcePath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("source is not a git repository: %s", sourcePath)
	}

	cmd := exec.Command(gitExecutable, "clone", "--local", "--single-branch", sourcePath, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	return nil
}

func setupOpsRepository(opsRepoPath string, cfg *config.Config) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(opsRepoPath); err != nil {
		return fmt.Errorf("failed to change to ops repository directory: %w", err)
	}

	gitCmd := func(args ...string) error {
		cmd := exec.Command(cfg.GitExecutable, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git %s failed: %w, output: %s", strings.Join(args, " "), err, string(output))
		}
		return nil
	}

	originalRemoteUrl, err := getOriginalRemoteUrl(cfg.GitExecutable)
	if err != nil {
		return fmt.Errorf("failed to get original remote URL: %w", err)
	}

	if err := gitCmd("remote", "set-url", "origin", originalRemoteUrl); err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	currentBranch, err := getCurrentBranch(cfg.GitExecutable)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch != cfg.TargetBranch {
		if err := gitCmd("checkout", "-b", cfg.TargetBranch); err != nil {
			if err := gitCmd("checkout", cfg.TargetBranch); err != nil {
				return fmt.Errorf("failed to create/checkout target branch: %w", err)
			}
		}
	}

	return nil
}

func getOriginalRemoteUrl(gitExecutable string) (string, error) {
	cmd := exec.Command(gitExecutable, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

func getCurrentBranch(gitExecutable string) (string, error) {
	cmd := exec.Command(gitExecutable, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}