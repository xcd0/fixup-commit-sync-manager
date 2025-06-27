# FixupCommitSyncManager - Implementation Guide

This document provides implementation details for the FixupCommitSyncManager, a Go-based tool for synchronizing source files between Dev and Ops repositories on Windows.

## Project Structure

```
FixupCommitSyncManager/
├── main.go                           # Entry point
├── go.mod                           # Go module definition
├── cmd/                             # Command implementations
│   ├── root.go                      # Root command and CLI setup
│   ├── init_config.go               # Interactive config generation
│   ├── validate_config.go           # Config validation
│   ├── init_vhdx.go                 # VHDX initialization
│   ├── mount_vhdx.go                # VHDX mounting
│   ├── unmount_vhdx.go              # VHDX unmounting
│   ├── snapshot_vhdx.go             # VHDX snapshot management
│   ├── sync.go                      # File synchronization
│   └── fixup.go                     # Fixup commit operations
├── internal/
│   ├── config/                      # Configuration management
│   │   ├── config.go                # Config struct and loading
│   │   └── config_test.go           # Config tests
│   ├── vhdx/                        # VHDX operations
│   │   ├── vhdx.go                  # VHDX manager
│   │   └── vhdx_test.go             # VHDX tests
│   ├── sync/                        # File synchronization
│   │   ├── sync.go                  # Sync logic
│   │   └── sync_test.go             # Sync tests
│   ├── fixup/                       # Fixup operations
│   │   ├── fixup.go                 # Fixup manager
│   │   └── fixup_test.go            # Fixup tests
│   ├── logger/                      # Logging utilities
│   │   ├── logger.go                # Logger implementation
│   │   └── logger_test.go           # Logger tests
│   ├── retry/                       # Retry mechanism
│   │   ├── retry.go                 # Retry logic
│   │   └── retry_test.go            # Retry tests
│   ├── notify/                      # Notification system
│   │   └── notify.go                # Slack notifications
│   └── utils/                       # Utility functions
│       └── utils.go                 # Common utilities
└── README.md                        # Project documentation
```

## Key Features Implemented

### 1. Configuration Management (`internal/config/`)
- HJSON-based configuration with comments
- Interactive wizard for config generation
- Comprehensive validation including paths and intervals
- Support for all required settings from specification

### 2. VHDX Management (`internal/vhdx/`)
- VHDX creation, mounting, and unmounting
- Snapshot management (create, list, rollback)
- PowerShell and diskpart integration for Windows
- Cross-platform testing support

### 3. File Synchronization (`internal/sync/`)
- Tracks changes in Dev repository
- Supports include/exclude patterns
- Preserves directory structure
- Automatic commit generation with templates

### 4. Fixup Operations (`internal/fixup/`)
- Automated fixup commits
- Autosquash rebase support
- Target branch management
- Continuous operation mode

### 5. Utilities (`internal/logger/`, `internal/retry/`, `internal/notify/`)
- Structured logging with levels and colors
- Retry mechanism for resilient operations
- Slack notification support
- Error handling and cleanup utilities

## Commands Implemented

### Core Commands
- `init-config`: Interactive configuration wizard
- `validate-config`: Configuration validation
- `sync`: File synchronization with continuous mode
- `fixup`: Fixup commit operations with continuous mode

### VHDX Commands
- `init-vhdx`: Initialize VHDX with repository clone
- `mount-vhdx`: Mount VHDX file
- `unmount-vhdx`: Unmount VHDX file
- `snapshot-vhdx create [name]`: Create snapshot
- `snapshot-vhdx list`: List snapshots
- `snapshot-vhdx rollback <name>`: Rollback to snapshot

### Global Flags
- `--config <path>`: Configuration file path
- `--dry-run`: Preview mode without changes
- `--verbose`: Detailed output
- `--continuous`: Continuous operation mode (sync/fixup)

## Build and Test

### Building
```bash
go build -o fixup-commit-sync-manager .
```

### Testing
```bash
go test ./...                    # All tests
go test ./internal/config        # Config tests only
go test -v ./internal/sync       # Verbose sync tests
```

### Cross-platform Considerations
- VHDX operations are Windows-specific but gracefully handle other platforms in tests
- Git operations work cross-platform
- File path handling uses filepath package for cross-platform compatibility

## Usage Examples

### Initialize Configuration
```bash
./fixup-commit-sync-manager init-config
```

### Validate Configuration
```bash
./fixup-commit-sync-manager validate-config --config my-config.hjson --verbose
```

### Sync Files Once
```bash
./fixup-commit-sync-manager sync --config my-config.hjson
```

### Continuous Sync
```bash
./fixup-commit-sync-manager sync --continuous --verbose
```

### VHDX Operations
```bash
./fixup-commit-sync-manager init-vhdx --config my-config.hjson
./fixup-commit-sync-manager snapshot-vhdx create backup-before-sync
./fixup-commit-sync-manager snapshot-vhdx list
```

## Configuration Example

```hjson
{
  // Repository paths (required)
  "devRepoPath": "C:\\path\\to\\dev-repo",
  "opsRepoPath": "C:\\path\\to\\ops-repo",
  
  // Sync settings
  "syncInterval": "5m",
  "includeExtensions": [".cpp", ".h", ".hpp"],
  "excludePatterns": ["bin/**", "obj/**"],
  
  // Fixup settings
  "fixupInterval": "1h",
  "targetBranch": "sync-branch",
  "baseBranch": "main",
  "autosquashEnabled": true,
  
  // VHDX settings
  "vhdxPath": "C:\\vhdx\\ops.vhdx",
  "mountPoint": "X:",
  "vhdxSize": "10GB",
  "encryptionEnabled": false,
  
  // Logging
  "logLevel": "INFO",
  "logFilePath": "C:\\logs\\sync.log",
  "verbose": false
}
```

## Implementation Status

✅ **Completed:**
1. Interactive config wizard with HJSON template generation
2. Configuration validation with comprehensive checks
3. VHDX creation, mounting, and snapshot management
4. File synchronization with pattern matching
5. Fixup commit operations with autosquash
6. Logging, retry mechanism, and error handling
7. Complete test coverage for core functionality
8. CLI with all specified subcommands

📝 **Notes:**
- VHDX operations require Windows environment for full functionality
- Git operations require git executable in PATH
- Slack notifications require webhook URL configuration
- Some tests may be skipped on non-Windows or non-Git environments

The implementation follows the specification requirements and provides a complete, testable solution for source file synchronization between Dev and Ops repositories.