package vhdx

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewVHDXManager(t *testing.T) {
	vhdxPath := "/path/to/test.vhdx"
	mountPoint := "X:"
	size := "10GB"
	encrypted := true

	manager := NewVHDXManager(vhdxPath, mountPoint, size, encrypted)

	if manager.VHDXPath != vhdxPath {
		t.Errorf("Expected VHDXPath %s, got %s", vhdxPath, manager.VHDXPath)
	}
	if manager.MountPoint != mountPoint {
		t.Errorf("Expected MountPoint %s, got %s", mountPoint, manager.MountPoint)
	}
	if manager.Size != size {
		t.Errorf("Expected Size %s, got %s", size, manager.Size)
	}
	if manager.Encrypted != encrypted {
		t.Errorf("Expected Encrypted %t, got %t", encrypted, manager.Encrypted)
	}
}

func TestParseSizeToMB(t *testing.T) {
	manager := &VHDXManager{}

	tests := []struct {
		size     string
		expected int
	}{
		{"10GB", 10240},
		{"5gb", 5120},
		{"1024MB", 1024},
		{"512mb", 512},
		{"invalid", 10240}, // default
		{"", 10240},        // default
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			manager.Size = tt.size
			result := manager.parseSizeToMB()
			if result != tt.expected {
				t.Errorf("parseSizeToMB() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"42", 42},
		{"", 0},
		{"abc", 0},
		{"12abc", 0},
		{"  123  ", 123},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateDiskpartScript(t *testing.T) {
	var vhdxPath string
	if runtime.GOOS == "windows" {
		vhdxPath = "C:\\test\\test.vhdx"
	} else {
		vhdxPath = "/test/test.vhdx"
	}

	manager := &VHDXManager{
		VHDXPath:   vhdxPath,
		MountPoint: "X:",
		Size:       "5GB",
	}

	script := manager.generateDiskpartScript()

	expectedStrings := []string{
		"maximum=5120",
		"assign letter=X",
	}

	for _, expected := range expectedStrings {
		if !contains(script, expected) {
			t.Errorf("Script should contain %q", expected)
		}
	}
}

func TestGetSnapshotPath(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")

	manager := &VHDXManager{
		VHDXPath: vhdxPath,
	}

	snapshotPath := manager.getSnapshotPath("test-snapshot")
	expectedPath := filepath.Join(tempDir, "snapshots", "test-snapshot.vhdx")

	if snapshotPath != expectedPath {
		t.Errorf("getSnapshotPath() = %s, want %s", snapshotPath, expectedPath)
	}
}

func TestListSnapshotsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")

	manager := &VHDXManager{
		VHDXPath: vhdxPath,
	}

	snapshots, err := manager.ListSnapshots()
	if err != nil {
		t.Errorf("ListSnapshots() failed: %v", err)
	}

	if len(snapshots) != 0 {
		t.Errorf("Expected empty snapshots list, got %d items", len(snapshots))
	}
}

func TestListSnapshotsWithFiles(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	snapshotDir := filepath.Join(tempDir, "snapshots")

	os.MkdirAll(snapshotDir, 0755)
	
	snapshotFiles := []string{"snap1.vhdx", "snap2.vhdx", "notavhdx.txt"}
	for _, file := range snapshotFiles {
		os.WriteFile(filepath.Join(snapshotDir, file), []byte("test"), 0644)
	}

	manager := &VHDXManager{
		VHDXPath: vhdxPath,
	}

	snapshots, err := manager.ListSnapshots()
	if err != nil {
		t.Errorf("ListSnapshots() failed: %v", err)
	}

	expectedCount := 2 // Only .vhdx files
	if len(snapshots) != expectedCount {
		t.Errorf("Expected %d snapshots, got %d", expectedCount, len(snapshots))
	}

	expectedSnapshots := map[string]bool{"snap1": true, "snap2": true}
	for _, snapshot := range snapshots {
		if !expectedSnapshots[snapshot] {
			t.Errorf("Unexpected snapshot: %s", snapshot)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}