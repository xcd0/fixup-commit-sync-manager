package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fixup-commit-sync-manager/internal/config"
)

type FileSyncer struct {
	cfg *config.Config
}

type SyncResult struct {
	FilesAdded    []string
	FilesModified []string
	FilesDeleted  []string
	CommitHash    string
}

func NewFileSyncer(cfg *config.Config) *FileSyncer {
	return &FileSyncer{cfg: cfg}
}

func (s *FileSyncer) Sync() (*SyncResult, error) {
	if s.isPaused() {
		return nil, fmt.Errorf("sync is paused by lock file: %s", s.cfg.PauseLockFile)
	}

	if err := s.validateRepositories(); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Dev側のカレントブランチを取得。
	devBranch, err := s.getDevCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get dev current branch: %w", err)
	}

	// Ops側を同じブランチに切り替え。
	if err := s.ensureOpsBranch(devBranch); err != nil {
		return nil, fmt.Errorf("failed to ensure ops branch: %w", err)
	}

	changes, err := s.detectChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to detect changes: %w", err)
	}

	if len(changes.FilesAdded)+len(changes.FilesModified)+len(changes.FilesDeleted) == 0 {
		return &SyncResult{}, nil
	}

	if err := s.applyChanges(changes); err != nil {
		return nil, fmt.Errorf("failed to apply changes: %w", err)
	}

	commitHash, err := s.commitChanges(changes)
	if err != nil {
		return nil, fmt.Errorf("failed to commit changes: %w", err)
	}

	changes.CommitHash = commitHash
	return changes, nil
}

func (s *FileSyncer) isPaused() bool {
	lockPath := filepath.Join(s.cfg.DevRepoPath, s.cfg.PauseLockFile)
	_, err := os.Stat(lockPath)
	return err == nil
}

func (s *FileSyncer) validateRepositories() error {
	devGitDir := filepath.Join(s.cfg.DevRepoPath, ".git")
	if _, err := os.Stat(devGitDir); err != nil {
		return fmt.Errorf("dev repository .git directory not found: %s", devGitDir)
	}

	opsGitDir := filepath.Join(s.cfg.OpsRepoPath, ".git")
	if _, err := os.Stat(opsGitDir); err != nil {
		return fmt.Errorf("ops repository .git directory not found: %s", opsGitDir)
	}

	return nil
}

func (s *FileSyncer) detectChanges() (*SyncResult, error) {
	result := &SyncResult{
		FilesAdded:    []string{},
		FilesModified: []string{},
		FilesDeleted:  []string{},
	}

	trackedChanges, err := s.getTrackedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked changes: %w", err)
	}

	newFiles, err := s.getNewFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get new files: %w", err)
	}

	for _, file := range trackedChanges {
		if s.shouldIncludeFile(file) {
			if s.fileExistsInOps(file) {
				result.FilesModified = append(result.FilesModified, file)
			} else {
				result.FilesDeleted = append(result.FilesDeleted, file)
			}
		}
	}

	for _, file := range newFiles {
		if s.shouldIncludeFile(file) {
			result.FilesAdded = append(result.FilesAdded, file)
		}
	}

	return result, nil
}

func (s *FileSyncer) getTrackedChanges() ([]string, error) {
	// 直前のコミットとの差分を取得。
	cmd := exec.Command(s.cfg.GitExecutable, "diff", "--name-only", "HEAD^")
	cmd.Dir = s.cfg.DevRepoPath
	output, err := cmd.Output()
	if err != nil {
		// HEAD^が存在しない場合（初回コミット）は全ファイルを対象とする。
		cmd = exec.Command(s.cfg.GitExecutable, "diff", "--name-only", "--cached")
		cmd.Dir = s.cfg.DevRepoPath
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git diff failed: %w", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

func (s *FileSyncer) getNewFiles() ([]string, error) {
	cmd := exec.Command(s.cfg.GitExecutable, "ls-files", "--others", "--exclude-standard")
	cmd.Dir = s.cfg.DevRepoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

func (s *FileSyncer) shouldIncludeFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, includeExt := range s.cfg.IncludeExtensions {
		if ext == strings.ToLower(includeExt) {
			return !s.isExcluded(filePath)
		}
	}

	for _, pattern := range s.cfg.IncludePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return !s.isExcluded(filePath)
		}
	}

	return false
}

func (s *FileSyncer) isExcluded(filePath string) bool {
	for _, pattern := range s.cfg.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
	}
	return false
}

func (s *FileSyncer) fileExistsInOps(filePath string) bool {
	opsFilePath := filepath.Join(s.cfg.OpsRepoPath, filePath)
	_, err := os.Stat(opsFilePath)
	return err == nil
}

func (s *FileSyncer) applyChanges(changes *SyncResult) error {
	for _, file := range changes.FilesAdded {
		if err := s.copyFileToOps(file); err != nil {
			return fmt.Errorf("failed to copy new file %s: %w", file, err)
		}
	}

	for _, file := range changes.FilesModified {
		if err := s.copyFileToOps(file); err != nil {
			return fmt.Errorf("failed to copy modified file %s: %w", file, err)
		}
	}

	for _, file := range changes.FilesDeleted {
		if err := s.deleteFileFromOps(file); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", file, err)
		}
	}

	return nil
}

func (s *FileSyncer) copyFileToOps(filePath string) error {
	srcPath := filepath.Join(s.cfg.DevRepoPath, filePath)
	dstPath := filepath.Join(s.cfg.OpsRepoPath, filePath)

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", srcPath)
	}

	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to destination file: %w", writeErr)
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}

func (s *FileSyncer) deleteFileFromOps(filePath string) error {
	opsFilePath := filepath.Join(s.cfg.OpsRepoPath, filePath)

	if _, err := os.Stat(opsFilePath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(opsFilePath); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", opsFilePath, err)
	}

	return nil
}

func (s *FileSyncer) commitChanges(changes *SyncResult) (string, error) {
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(s.cfg.OpsRepoPath); err != nil {
		return "", fmt.Errorf("failed to change to ops repository: %w", err)
	}

	if err := s.gitAddChanges(); err != nil {
		return "", fmt.Errorf("failed to add changes: %w", err)
	}

	commitMsg := s.generateCommitMessage(changes)
	if err := s.gitCommit(commitMsg); err != nil {
		return "", fmt.Errorf("failed to commit changes: %w", err)
	}

	return s.getLastCommitHash()
}

func (s *FileSyncer) gitAddChanges() error {
	cmd := exec.Command(s.cfg.GitExecutable, "add", "-A")
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (s *FileSyncer) gitCommit(message string) error {
	args := []string{"commit", "-m", message}

	if s.cfg.AuthorName != "" && s.cfg.AuthorEmail != "" {
		author := fmt.Sprintf("%s <%s>", s.cfg.AuthorName, s.cfg.AuthorEmail)
		args = append(args, "--author", author)
	}

	cmd := exec.Command(s.cfg.GitExecutable, args...)
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w, output: %s", err, string(output))
	}
	return nil
}

func (s *FileSyncer) getLastCommitHash() (string, error) {
	cmd := exec.Command(s.cfg.GitExecutable, "rev-parse", "HEAD")
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *FileSyncer) generateCommitMessage(changes *SyncResult) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	message := s.cfg.CommitTemplate
	message = strings.ReplaceAll(message, "${timestamp}", timestamp)

	if changes.CommitHash != "" {
		message = strings.ReplaceAll(message, "${hash}", changes.CommitHash[:8])
	} else {
		message = strings.ReplaceAll(message, "${hash}", "pending")
	}

	totalFiles := len(changes.FilesAdded) + len(changes.FilesModified) + len(changes.FilesDeleted)
	summary := fmt.Sprintf(" (%d files: +%d ~%d -%d)",
		totalFiles, len(changes.FilesAdded), len(changes.FilesModified), len(changes.FilesDeleted))

	return message + summary
}

// getDevCurrentBranch はDev側のカレントブランチを取得する。
func (s *FileSyncer) getDevCurrentBranch() (string, error) {
	cmd := exec.Command(s.cfg.GitExecutable, "branch", "--show-current")
	cmd.Dir = s.cfg.DevRepoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch from dev repo: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ensureOpsBranch はOps側を指定されたブランチに切り替える。
func (s *FileSyncer) ensureOpsBranch(targetBranch string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(s.cfg.OpsRepoPath); err != nil {
		return fmt.Errorf("failed to change to ops repository: %w", err)
	}

	// 現在のブランチを確認。
	currentBranch, err := s.getOpsCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current ops branch: %w", err)
	}

	// 既に目的のブランチにいる場合はスキップ。
	if currentBranch == targetBranch {
		return nil
	}

	// ブランチの存在確認。
	if err := s.ensureBranchExists(targetBranch); err != nil {
		return fmt.Errorf("failed to ensure branch exists: %w", err)
	}

	// ブランチ切り替え。
	if err := s.gitCheckout(targetBranch); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", targetBranch, err)
	}

	return nil
}

// getOpsCurrentBranch はOps側のカレントブランチを取得する。
func (s *FileSyncer) getOpsCurrentBranch() (string, error) {
	cmd := exec.Command(s.cfg.GitExecutable, "branch", "--show-current")
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch from ops repo: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ensureBranchExists は指定されたブランチが存在することを確認し、必要に応じて作成する。
func (s *FileSyncer) ensureBranchExists(branchName string) error {
	// ローカルブランチの存在確認。
	cmd := exec.Command(s.cfg.GitExecutable, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = s.cfg.OpsRepoPath
	if err := cmd.Run(); err == nil {
		return nil // ブランチが存在する。
	}

	// リモートブランチの存在確認。
	cmd = exec.Command(s.cfg.GitExecutable, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branchName)
	cmd.Dir = s.cfg.OpsRepoPath
	if err := cmd.Run(); err == nil {
		// リモートブランチから作成。
		return s.gitCreateBranchFromRemote(branchName)
	}

	// ブランチが存在しない場合は新規作成。
	return s.gitCreateBranch(branchName)
}

// gitCreateBranch は新しいブランチを作成する。
func (s *FileSyncer) gitCreateBranch(branchName string) error {
	cmd := exec.Command(s.cfg.GitExecutable, "checkout", "-b", branchName)
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w, output: %s", branchName, err, string(output))
	}
	return nil
}

// gitCreateBranchFromRemote はリモートブランチから新しいローカルブランチを作成する。
func (s *FileSyncer) gitCreateBranchFromRemote(branchName string) error {
	cmd := exec.Command(s.cfg.GitExecutable, "checkout", "-b", branchName, "origin/"+branchName)
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s from remote: %w, output: %s", branchName, err, string(output))
	}
	return nil
}

// gitCheckout は指定されたブランチに切り替える。
func (s *FileSyncer) gitCheckout(branchName string) error {
	cmd := exec.Command(s.cfg.GitExecutable, "checkout", branchName)
	cmd.Dir = s.cfg.OpsRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w, output: %s", branchName, err, string(output))
	}
	return nil
}
