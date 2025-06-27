package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"fixup-commit-sync-manager/internal/config"
)

func TestValidatePaths(t *testing.T) {
	tempDir := t.TempDir()
	devPath := filepath.Join(tempDir, "dev")
	opsPath := filepath.Join(tempDir, "ops")

	os.MkdirAll(devPath, 0755)
	os.MkdirAll(opsPath, 0755)

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid paths",
			cfg: &config.Config{
				DevRepoPath: devPath,
				OpsRepoPath: opsPath,
			},
			wantErr: false,
		},
		{
			name: "relative dev path",
			cfg: &config.Config{
				DevRepoPath: "relative/path",
				OpsRepoPath: opsPath,
			},
			wantErr: true,
		},
		{
			name: "relative ops path",
			cfg: &config.Config{
				DevRepoPath: devPath,
				OpsRepoPath: "relative/path",
			},
			wantErr: true,
		},
		{
			name: "same paths",
			cfg: &config.Config{
				DevRepoPath: devPath,
				OpsRepoPath: devPath,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaths(tt.cfg, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePaths() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVHDXConfig(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "no VHDX config",
			cfg: &config.Config{
				VHDXPath: "",
			},
			wantErr: false,
		},
		{
			name: "valid VHDX config",
			cfg: &config.Config{
				VHDXPath:   vhdxPath,
				MountPoint: "X:",
			},
			wantErr: false,
		},
		{
			name: "relative VHDX path",
			cfg: &config.Config{
				VHDXPath:   "relative/path.vhdx",
				MountPoint: "X:",
			},
			wantErr: true,
		},
		{
			name: "missing mount point",
			cfg: &config.Config{
				VHDXPath:   vhdxPath,
				MountPoint: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent VHDX directory",
			cfg: &config.Config{
				VHDXPath:   "/non/existent/path/test.vhdx",
				MountPoint: "X:",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVHDXConfig(tt.cfg, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVHDXConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
