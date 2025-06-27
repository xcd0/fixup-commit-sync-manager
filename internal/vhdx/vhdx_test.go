package vhdx

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// ===== 新規追加テスト =====

// TestNewManager は既存の互換性コンストラクタのテスト。
func TestNewManager(t *testing.T) {
	vhdxPath := "/test/path.vhdx"
	mountPoint := "Y:"
	
	manager := NewManager(vhdxPath, mountPoint)
	
	if manager.VHDXPath != vhdxPath {
		t.Errorf("Expected VHDXPath %s, got %s", vhdxPath, manager.VHDXPath)
	}
	if manager.MountPoint != mountPoint {
		t.Errorf("Expected MountPoint %s, got %s", mountPoint, manager.MountPoint)
	}
	if manager.Size != "10GB" {
		t.Errorf("Expected default Size 10GB, got %s", manager.Size)
	}
	if manager.Encrypted != false {
		t.Errorf("Expected default Encrypted false, got %t", manager.Encrypted)
	}
}

// TestParseSizeToBytes はgo-winio用のバイト変換をテスト。
func TestParseSizeToBytes(t *testing.T) {
	manager := &VHDXManager{}

	tests := []struct {
		size     string
		expected uint64
	}{
		{"10GB", 10 * 1024 * 1024 * 1024},
		{"5gb", 5 * 1024 * 1024 * 1024},
		{"1024MB", 1024 * 1024 * 1024},
		{"512mb", 512 * 1024 * 1024},
		{"2TB", 2 * 1024 * 1024 * 1024 * 1024},
		{"1tb", 1024 * 1024 * 1024 * 1024},
		{"invalid", 10 * 1024 * 1024 * 1024}, // default 10GB
		{"", 10 * 1024 * 1024 * 1024},        // default 10GB
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			manager.Size = tt.size
			result := manager.parseSizeToBytes()
			if result != tt.expected {
				t.Errorf("parseSizeToBytes() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestCreateVHDX はVHDX作成機能のテスト。
func TestCreateVHDX(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// 既存ファイルのテスト。
	existingFile, err := os.Create(vhdxPath)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}
	existingFile.Close()
	
	err = manager.CreateVHDX()
	if err == nil {
		t.Error("CreateVHDX() should fail for existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention file already exists: %v", err)
	}
}

// TestCreate はCreate関数のテスト。
func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "new.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// 新しいサイズと暗号化設定でテスト。
	err := manager.Create("20GB", true)
	
	// 非Windows環境では失敗することが期待される。
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Create() should fail on non-Windows environment")
		}
	}
	
	// サイズと暗号化設定が更新されることを確認。
	if manager.Size != "20GB" {
		t.Errorf("Expected Size to be updated to 20GB, got %s", manager.Size)
	}
	if manager.Encrypted != true {
		t.Errorf("Expected Encrypted to be updated to true, got %t", manager.Encrypted)
	}
}

// TestMountVHDX はマウント機能のテスト。
func TestMountVHDX(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	mountPoint := "Z:"
	
	manager := NewManager(vhdxPath, mountPoint)
	
	// 非Windows環境ではエラーが期待される。
	err := manager.Mount()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("Mount() should fail on non-Windows environment")
		}
	}
	
	// MountVHDX関数も同様。
	err = manager.MountVHDX()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("MountVHDX() should fail on non-Windows environment")
		}
	}
}

// TestUnmountVHDX はアンマウント機能のテスト。
func TestUnmountVHDX(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	mountPoint := "Z:"
	
	manager := NewManager(vhdxPath, mountPoint)
	
	// 非Windows環境ではエラーが期待される。
	err := manager.UnmountVHDX()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("UnmountVHDX() should fail on non-Windows environment")
		}
	}
}

// TestCreateSnapshot はスナップショット作成のテスト。
func TestCreateSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// 既存のVHDXファイルを作成。
	if err := os.WriteFile(vhdxPath, []byte("dummy vhdx"), 0644); err != nil {
		t.Fatalf("Failed to create dummy VHDX: %v", err)
	}
	
	// 名前を指定しないスナップショット作成。
	err := manager.CreateSnapshot("")
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("CreateSnapshot() should fail on non-Windows environment")
		}
	}
	
	// 名前を指定したスナップショット作成。
	err = manager.CreateSnapshot("test-snapshot")
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("CreateSnapshot() should fail on non-Windows environment")
		}
	}
}

// TestRollbackToSnapshot はスナップショットロールバックのテスト。
func TestRollbackToSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// 存在しないスナップショットのテスト。
	err := manager.RollbackToSnapshot("nonexistent")
	if err == nil {
		t.Error("RollbackToSnapshot() should fail for non-existent snapshot")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention snapshot not found: %v", err)
	}
}

// TestIsMounted はマウント状態確認のテスト。
func TestIsMounted(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	// 存在しないマウントポイントのテスト。
	manager := NewManager(vhdxPath, "/nonexistent/mount")
	if manager.isMounted() {
		t.Error("isMounted() should return false for non-existent mount point")
	}
	
	// 存在するディレクトリのテスト。
	manager.MountPoint = tempDir
	if !manager.isMounted() {
		t.Error("isMounted() should return true for existing directory")
	}
}

// TestExecuteDiskpartScript はdiskpartスクリプト実行のテスト。
func TestExecuteDiskpartScript(t *testing.T) {
	manager := &VHDXManager{}
	
	// 無効なスクリプトでテスト。
	err := manager.executeDiskpartScript("invalid script")
	if err == nil {
		t.Error("executeDiskpartScript() should fail for invalid script")
	}
}

// TestCopyFile はファイルコピーのテスト。
func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	
	// ソースファイルを作成。
	testContent := []byte("test file content")
	if err := os.WriteFile(srcPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// ファイルをコピー。
	err := copyFile(srcPath, dstPath)
	if err != nil {
		t.Errorf("copyFile() failed: %v", err)
	}
	
	// コピーされたファイルの内容を確認。
	copiedContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	
	if string(copiedContent) != string(testContent) {
		t.Errorf("Copied content does not match. Expected %s, got %s", testContent, copiedContent)
	}
	
	// 存在しないソースファイルのテスト。
	err = copyFile("/nonexistent/source.txt", dstPath)
	if err == nil {
		t.Error("copyFile() should fail for non-existent source file")
	}
}

// TestPlatformSpecificFunctions はプラットフォーム固有機能のテスト。
func TestPlatformSpecificFunctions(t *testing.T) {
	manager := &VHDXManager{}
	
	// hasGoWinioSupport のテスト。
	hasSupport := manager.hasGoWinioSupport()
	
	if runtime.GOOS == "windows" {
		if !hasSupport {
			t.Error("hasGoWinioSupport() should return true on Windows")
		}
	} else {
		if hasSupport {
			t.Error("hasGoWinioSupport() should return false on non-Windows")
		}
	}
	
	// プラットフォーム固有の実装テスト。
	err := manager.createVHDXWithGoWinio()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("createVHDXWithGoWinio() should fail on non-Windows")
		}
		if !strings.Contains(err.Error(), "only supported on Windows") {
			t.Errorf("Error should mention Windows support: %v", err)
		}
	}
	
	err = manager.mountVHDXWithGoWinio()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("mountVHDXWithGoWinio() should fail on non-Windows")
		}
	}
	
	err = manager.unmountVHDXWithGoWinio()
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("unmountVHDXWithGoWinio() should fail on non-Windows")
		}
	}
	
	err = manager.createSnapshotWithGoWinio("test-snapshot.vhdx")
	if runtime.GOOS != "windows" {
		if err == nil {
			t.Error("createSnapshotWithGoWinio() should fail on non-Windows")
		}
	}
}

// TestVirtualDiskType はVirtualDisk型のテスト。
func TestVirtualDiskType(t *testing.T) {
	var vd VirtualDisk
	
	// Close関数のテスト。
	err := vd.Close()
	if err != nil {
		t.Errorf("VirtualDisk.Close() should not fail: %v", err)
	}
}

// TestAssignDriveLetter はドライブレター割り当てのテスト（非Windows環境では実行されない）。
func TestAssignDriveLetter(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("assignDriveLetter test is Windows-only")
	}
	
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// PowerShellが利用できない環境ではエラーが期待される。
	err := manager.assignDriveLetter()
	// PowerShellが利用できない場合はエラーが発生することを確認。
	// テスト環境では実際のVHDXがないためエラーになることが多い。
	t.Logf("assignDriveLetter result: %v", err)
}

// TestCreateSnapshotWithPowerShell はPowerShell方式のスナップショット作成テスト。
func TestCreateSnapshotWithPowerShell(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	snapshotPath := filepath.Join(tempDir, "snapshot.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// PowerShellが利用できない環境ではエラーが期待される。
	err := manager.createSnapshotWithPowerShell(snapshotPath)
	if err == nil {
		t.Error("createSnapshotWithPowerShell() should fail in test environment")
	}
}

// TestInitializeAndFormatVHDX はVHDX初期化のテスト。
func TestInitializeAndFormatVHDX(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("initializeAndFormatVHDX test is Windows-only")
	}
	
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// PowerShellが利用できない環境やVHDXファイルがない場合はエラーが期待される。
	err := manager.initializeAndFormatVHDX()
	// テスト環境では実際のVHDXがないためエラーになることが多い。
	t.Logf("initializeAndFormatVHDX result: %v", err)
}
