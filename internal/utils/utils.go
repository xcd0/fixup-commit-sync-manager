package utils

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func HandleInterrupt(cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal, cleaning up...")
		if cleanup != nil {
			cleanup()
		}
		os.Exit(0)
	}()
}

func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

func SanitizeFilePath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimSpace(path)
	return path
}

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func StringSliceContains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func RemoveFromStringSlice(slice []string, str string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != str {
			result = append(result, item)
		}
	}
	return result
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func GetRelativePath(basePath, targetPath string) (string, error) {
	return filepath.Rel(basePath, targetPath)
}

func CleanupTempFiles(patterns []string) {
	tempDir := os.TempDir()
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(tempDir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			os.Remove(match)
		}
	}
}

func ValidateRequiredPaths(paths map[string]string) error {
	for name, path := range paths {
		if path == "" {
			return fmt.Errorf("%s path is required", name)
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("%s path must be absolute: %s", name, path)
		}
	}
	return nil
}

func CreateLockFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func RemoveLockFile(path string) error {
	if IsFileExists(path) {
		return os.Remove(path)
	}
	return nil
}

func TimedOperation(operation func() error, timeout time.Duration) error {
	done := make(chan error, 1)
	
	go func() {
		done <- operation()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}