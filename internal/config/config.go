package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hjson/hjson-go/v4"
)

type NotifyConfig struct {
	SlackWebhookURL string `json:"slackWebhookUrl,omitempty"`
}

type Config struct {
	DevRepoPath       string        `json:"devRepoPath"`
	OpsRepoPath       string        `json:"opsRepoPath"`
	IncludeExtensions []string      `json:"includeExtensions"`
	IncludePatterns   []string      `json:"includePatterns"`
	ExcludePatterns   []string      `json:"excludePatterns"`
	SyncInterval      string        `json:"syncInterval"`
	PauseLockFile     string        `json:"pauseLockFile"`
	GitExecutable     string        `json:"gitExecutable"`
	CommitTemplate    string        `json:"commitTemplate"`
	AuthorName        string        `json:"authorName,omitempty"`
	AuthorEmail       string        `json:"authorEmail,omitempty"`
	FixupInterval     string        `json:"fixupInterval"`
	FixupMsgPrefix    string        `json:"fixupMessagePrefix"`
	AutosquashEnabled bool          `json:"autosquashEnabled"`
	TargetBranch      string        `json:"targetBranch"`
	BaseBranch        string        `json:"baseBranch"`
	MaxRetries        int           `json:"maxRetries"`
	RetryDelay        string        `json:"retryDelay"`
	LogLevel          string        `json:"logLevel"`
	LogFilePath       string        `json:"logFilePath"`
	NotifyOnError     *NotifyConfig `json:"notifyOnError,omitempty"`
	DryRun            bool          `json:"dryRun"`
	Verbose           bool          `json:"verbose"`
	VHDXPath          string        `json:"vhdxPath,omitempty"`
	VHDXSize          string        `json:"vhdxSize"`
	MountPoint        string        `json:"mountPoint,omitempty"`
	EncryptionEnabled bool          `json:"encryptionEnabled"`
}

func DefaultConfig() *Config {
	return &Config{
		IncludeExtensions: []string{".cpp", ".h", ".hpp"},
		IncludePatterns:   []string{},
		ExcludePatterns:   []string{},
		SyncInterval:      "5m",
		PauseLockFile:     ".sync-paused",
		GitExecutable:     "git",
		CommitTemplate:    "Auto-sync: ${timestamp} @ ${hash}",
		FixupInterval:     "1h",
		FixupMsgPrefix:    "fixup! ",
		AutosquashEnabled: true,
		TargetBranch:      "sync-branch",
		BaseBranch:        "main",
		MaxRetries:        3,
		RetryDelay:        "30s",
		LogLevel:          "INFO",
		LogFilePath:       "./sync.log",
		DryRun:            false,
		Verbose:           false,
		VHDXSize:          "10GB",
		EncryptionEnabled: false,
	}
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonData interface{}
	if err := hjson.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse HJSON: %w", err)
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func (c *Config) GetSyncIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(c.SyncInterval)
}

func (c *Config) GetFixupIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(c.FixupInterval)
}

func (c *Config) GetRetryDelayDuration() (time.Duration, error) {
	return time.ParseDuration(c.RetryDelay)
}

func (c *Config) Validate() error {
	if c.DevRepoPath == "" {
		return fmt.Errorf("devRepoPath is required")
	}
	if c.OpsRepoPath == "" {
		return fmt.Errorf("opsRepoPath is required")
	}

	if _, err := c.GetSyncIntervalDuration(); err != nil {
		return fmt.Errorf("invalid syncInterval: %w", err)
	}
	if _, err := c.GetFixupIntervalDuration(); err != nil {
		return fmt.Errorf("invalid fixupInterval: %w", err)
	}
	if _, err := c.GetRetryDelayDuration(); err != nil {
		return fmt.Errorf("invalid retryDelay: %w", err)
	}

	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid logLevel: must be one of DEBUG, INFO, WARN, ERROR")
	}

	return nil
}