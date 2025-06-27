# FixupCommitSyncManager - Implementation Guide

This document provides implementation details for the FixupCommitSyncManager, a Go-based tool for synchronizing source files between Dev and Ops repositories with dynamic branch tracking.

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
- **Dynamic branch tracking**: No manual branch configuration required

### 2. VHDX Management (`internal/vhdx/`)
- VHDX creation, mounting, and unmounting
- Snapshot management (create, list, rollback)
- PowerShell and diskpart integration for Windows
- Cross-platform testing support

### 3. File Synchronization (`internal/sync/`)
- **Dynamic branch tracking**: Automatically detects Dev repository's current branch
- **Auto branch switching**: Ops repository follows Dev repository's branch
- **Branch creation**: Creates missing branches locally or from remote
- Tracks changes from previous commit (HEAD^) vs current state
- Supports include/exclude patterns
- Preserves directory structure
- Automatic commit generation with templates

### 4. Fixup Operations (`internal/fixup/`)
- **Dynamic branch tracking**: Follows Dev repository's current branch
- Automated fixup commits against previous commit
- Autosquash rebase support
- Continuous operation mode
- No fixed target/base branch dependency

### 5. Utilities (`internal/logger/`, `internal/retry/`, `internal/notify/`)
- Structured logging with levels and colors
- Retry mechanism for resilient operations
- Slack notification support
- Error handling and cleanup utilities

## Commands Implemented

### Core Commands
- `init-config`: Interactive configuration wizard (no branch configuration needed)
- `validate-config`: Configuration validation
- `sync`: File synchronization with dynamic branch tracking and continuous mode
- `fixup`: Fixup commit operations with dynamic branch tracking and continuous mode

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
  "autosquashEnabled": true,
  // Note: Branch settings are now dynamic - automatically tracks Dev repository's current branch
  
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
1. **Dynamic branch tracking**: Automatically follows Dev repository's current branch
2. Interactive config wizard with HJSON template generation (no branch config needed)
3. Configuration validation with comprehensive checks
4. VHDX creation, mounting, and snapshot management
5. File synchronization with dynamic branch switching and pattern matching
6. Fixup commit operations with dynamic branch tracking and autosquash
7. Logging, retry mechanism, and error handling
8. Complete test coverage for core functionality including dynamic branch features
9. CLI with all specified subcommands

## Dynamic Branch Tracking Features

### Sync Process Flow:
1. Detects Dev repository's current branch (e.g., `feature-abc`)
2. Switches Ops repository to the same branch (`feature-abc`)
3. Creates branch if it doesn't exist (locally or from remote)
4. Compares Dev's previous commit (HEAD^) with current state
5. Syncs differences to Ops repository on the same branch

### Fixup Process Flow:
1. Detects Dev repository's current branch
2. Switches Ops repository to the same branch
3. Creates fixup commits against the previous commit
4. Applies autosquash rebase if enabled

### Branch Management:
- **Automatic creation**: Creates missing branches locally or from remote origin
- **No configuration**: No manual branch specification needed
- **Dynamic switching**: Always follows Dev repository's current state
- **Backward compatibility**: Works with existing repositories

📝 **Notes:**
- VHDX operations require Windows environment for full functionality
- Git operations require git executable in PATH
- Slack notifications require webhook URL configuration
- Some tests may be skipped on non-Windows or non-Git environments
- Dynamic branch tracking works with any branch name - no restrictions

The implementation provides a robust, dynamic solution for source file synchronization that automatically adapts to the developer's workflow without manual branch configuration.