package fixup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"fixup-commit-sync-manager/internal/config"
)

type FixupManager struct {
	cfg *config.Config
}

type FixupResult struct {
	CommitHash      string
	FixupCommitHash string
	FilesModified   int
	Success         bool
}

func NewFixupManager(cfg *config.Config) *FixupManager {
	return &FixupManager{cfg: cfg}
}

func (f *FixupManager) RunFixup() (*FixupResult, error) {
	if err := f.validateRepository(); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	hasChanges, err := f.hasUncommittedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if !hasChanges {
		return &FixupResult{Success: true}, nil
	}

	baseCommit, err := f.getBaseCommit()
	if err != nil {
		return nil, fmt.Errorf("failed to get base commit: %w", err)
	}

	modifiedFiles, err := f.getModifiedFilesCount()
	if err != nil {
		return nil, fmt.Errorf("failed to count modified files: %w", err)
	}

	if err := f.gitAddAll(); err != nil {
		return nil, fmt.Errorf("failed to add changes: %w", err)
	}

	fixupHash, err := f.gitFixupCommit(baseCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to create fixup commit: %w", err)
	}

	if f.cfg.AutosquashEnabled {
		if err := f.gitRebaseAutosquash(); err != nil {
			return nil, fmt.Errorf("failed to perform autosquash rebase: %w", err)
		}
	}

	return &FixupResult{
		CommitHash:      baseCommit,
		FixupCommitHash: fixupHash,
		FilesModified:   modifiedFiles,
		Success:         true,
	}, nil
}

func (f *FixupManager) validateRepository() error {
	opsGitDir := f.cfg.OpsRepoPath + "/.git"
	if _, err := os.Stat(opsGitDir); err != nil {
		return fmt.Errorf("ops repository .git directory not found: %s", opsGitDir)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(f.cfg.OpsRepoPath); err != nil {
		return fmt.Errorf("failed to change to ops repository: %w", err)
	}

	if err := f.ensureOnTargetBranch(); err != nil {
		return fmt.Errorf("failed to ensure on target branch: %w", err)
	}

	return nil
}

func (f *FixupManager) ensureOnTargetBranch() error {
	currentBranch, err := f.getCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch != f.cfg.TargetBranch {
		if err := f.gitCheckoutBranch(f.cfg.TargetBranch); err != nil {
			return fmt.Errorf("failed to checkout target branch %s: %w", f.cfg.TargetBranch, err)
		}
	}

	return nil
}

func (f *FixupManager) getCurrentBranch() (string, error) {
	cmd := exec.Command(f.cfg.GitExecutable, "branch", "--show-current")
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (f *FixupManager) gitCheckoutBranch(branch string) error {
	cmd := exec.Command(f.cfg.GitExecutable, "checkout", branch)
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (f *FixupManager) hasUncommittedChanges() (bool, error) {
	cmd := exec.Command(f.cfg.GitExecutable, "status", "--porcelain")
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

func (f *FixupManager) getBaseCommit() (string, error) {
	cmd := exec.Command(f.cfg.GitExecutable, "merge-base", "HEAD", f.cfg.BaseBranch)
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command(f.cfg.GitExecutable, "rev-parse", "HEAD~1")
		cmd.Dir = f.cfg.OpsRepoPath
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get base commit: %w", err)
		}
	}

	return strings.TrimSpace(string(output)), nil
}

func (f *FixupManager) getModifiedFilesCount() (int, error) {
	cmd := exec.Command(f.cfg.GitExecutable, "diff", "--name-only", "--cached")
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get modified files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	return len(lines), nil
}

func (f *FixupManager) gitAddAll() error {
	cmd := exec.Command(f.cfg.GitExecutable, "add", "-u")
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (f *FixupManager) gitFixupCommit(baseCommit string) (string, error) {
	fixupTarget := fmt.Sprintf("%s", baseCommit)
	commitMsg := f.generateFixupMessage(baseCommit)

	args := []string{"commit", "--fixup", fixupTarget, "-m", commitMsg}
	
	if f.cfg.AuthorName != "" && f.cfg.AuthorEmail != "" {
		author := fmt.Sprintf("%s <%s>", f.cfg.AuthorName, f.cfg.AuthorEmail)
		args = append(args, "--author", author)
	}

	cmd := exec.Command(f.cfg.GitExecutable, args...)
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git fixup commit failed: %w, output: %s", err, string(output))
	}

	return f.getLastCommitHash()
}

func (f *FixupManager) gitRebaseAutosquash() error {
	baseCommit, err := f.getBaseCommit()
	if err != nil {
		return fmt.Errorf("failed to get base commit for rebase: %w", err)
	}

	cmd := exec.Command(f.cfg.GitExecutable, "rebase", "--autosquash", "--interactive", baseCommit)
	cmd.Dir = f.cfg.OpsRepoPath
	cmd.Env = append(os.Environ(), "GIT_EDITOR=true")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git rebase autosquash failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (f *FixupManager) getLastCommitHash() (string, error) {
	cmd := exec.Command(f.cfg.GitExecutable, "rev-parse", "HEAD")
	cmd.Dir = f.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (f *FixupManager) generateFixupMessage(baseCommit string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	message := f.cfg.FixupMsgPrefix + "Automated fixup"
	if baseCommit != "" {
		message += fmt.Sprintf(" for %s", baseCommit[:8])
	}
	message += fmt.Sprintf(" @ %s", timestamp)
	
	return message
}

func (f *FixupManager) RunContinuousFixup() error {
	interval, err := f.cfg.GetFixupIntervalDuration()
	if err != nil {
		return fmt.Errorf("invalid fixup interval: %w", err)
	}

	fmt.Printf("Starting continuous fixup with interval: %s\n", f.cfg.FixupInterval)
	fmt.Printf("Ops Repository: %s\n", f.cfg.OpsRepoPath)
	fmt.Printf("Target Branch: %s\n", f.cfg.TargetBranch)
	fmt.Printf("Base Branch: %s\n", f.cfg.BaseBranch)
	fmt.Println("Press Ctrl+C to stop")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if f.cfg.Verbose {
				fmt.Printf("\n[%s] Starting fixup operation...\n", time.Now().Format("15:04:05"))
			}

			if f.cfg.DryRun {
				fmt.Printf("[%s] [DRY RUN] Would perform fixup operation\n", time.Now().Format("15:04:05"))
				continue
			}

			result, err := f.RunFixup()
			if err != nil {
				fmt.Printf("[%s] Fixup failed: %v\n", time.Now().Format("15:04:05"), err)
				continue
			}

			if result.FilesModified == 0 {
				if f.cfg.Verbose {
					fmt.Printf("[%s] No changes to fixup\n", time.Now().Format("15:04:05"))
				}
				continue
			}

			fmt.Printf("[%s] ✓ Fixup completed - %d files modified", 
				time.Now().Format("15:04:05"), result.FilesModified)
			
			if result.FixupCommitHash != "" {
				fmt.Printf(" Commit: %s", result.FixupCommitHash[:8])
			}
			fmt.Println()
		}
	}
}