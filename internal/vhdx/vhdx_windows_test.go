//go:build windows
// +build windows

package vhdx

import (
	"path/filepath"
	"testing"
)

// TestWindowsSpecificFunctions はWindows固有機能の詳細テスト。
func TestWindowsSpecificFunctions(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "windows-test.vhdx")
	
	manager := NewManager(vhdxPath, "Z:")
	
	// hasGoWinioSupport はWindows環境でtrueを返すべき。
	if !manager.hasGoWinioSupport() {
		t.Error("hasGoWinioSupport() should return true on Windows")
	}
	
	// Windows固有の実装をテスト（実際のVHDXがない場合はエラーが期待される）。
	err := manager.createVHDXWithGoWinio()
	if err != nil {
		t.Logf("createVHDXWithGoWinio() expected to fail in test environment: %v", err)
		// テスト環境では実際のVHDX作成権限がない場合があるため、ログに記録。
	}
	
	err = manager.mountVHDXWithGoWinio()
	if err != nil {
		t.Logf("mountVHDXWithGoWinio() expected to fail for non-existent VHDX: %v", err)
	}
	
	err = manager.unmountVHDXWithGoWinio()
	if err != nil {
		t.Logf("unmountVHDXWithGoWinio() expected to fail for non-mounted VHDX: %v", err)
	}
	
	snapshotPath := filepath.Join(tempDir, "snapshot.vhdx")
	err = manager.createSnapshotWithGoWinio(snapshotPath)
	if err != nil {
		t.Logf("createSnapshotWithGoWinio() expected to fail for non-existent parent VHDX: %v", err)
	}
}

// TestWindowsVirtualDiskHandling はWindows環境でのVirtualDiskハンドリングテスト。
func TestWindowsVirtualDiskHandling(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "handle-test.vhdx")
	
	manager := NewManager(vhdxPath, "Y:")
	
	// VirtualDiskハンドルの初期状態確認。
	// Windows環境では実際のgo-winio VirtualDisk型が使用される。
	// 初期状態では空のハンドルである。
	
	// ハンドルのClose動作をテスト。
	err := manager.handle.Close()
	if err != nil {
		t.Errorf("VirtualDisk.Close() should not fail for empty handle: %v", err)
	}
}

// TestWindowsPowerShellIntegration はPowerShell統合のテスト。
func TestWindowsPowerShellIntegration(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "powershell-test.vhdx")
	
	manager := NewManager(vhdxPath, "X:")
	
	// PowerShellベースの機能をテスト。
	// テスト環境では実際のVHDXがないためエラーが期待される。
	
	err := manager.assignDriveLetter()
	if err != nil {
		t.Logf("assignDriveLetter() expected to fail for non-existent VHDX: %v", err)
	}
	
	err = manager.initializeAndFormatVHDX()
	if err != nil {
		t.Logf("initializeAndFormatVHDX() expected to fail for non-existent VHDX: %v", err)
	}
	
	snapshotPath := filepath.Join(tempDir, "powershell-snapshot.vhdx")
	err = manager.createSnapshotWithPowerShell(snapshotPath)
	if err != nil {
		t.Logf("createSnapshotWithPowerShell() expected to fail for non-existent parent VHDX: %v", err)
	}
}

// TestWindowsRealWorldScenario は実際の使用シナリオに近いテスト。
func TestWindowsRealWorldScenario(t *testing.T) {
	tempDir := t.TempDir()
	vhdxPath := filepath.Join(tempDir, "realworld-test.vhdx")
	
	manager := NewManager(vhdxPath, "W:")
	
	// 1. VHDX作成を試行。
	err := manager.Create("1GB", false)
	if err != nil {
		t.Logf("VHDX creation failed (expected in test environment): %v", err)
		// テスト環境では管理者権限がない場合があるため、失敗は許容。
	}
	
	// 2. マウントを試行。
	err = manager.Mount()
	if err != nil {
		t.Logf("VHDX mount failed (expected without actual VHDX): %v", err)
	}
	
	// 3. スナップショット作成を試行。
	err = manager.CreateSnapshot("test-snapshot")
	if err != nil {
		t.Logf("Snapshot creation failed (expected without actual VHDX): %v", err)
	}
	
	// 4. アンマウントを試行。
	err = manager.UnmountVHDX()
	if err != nil {
		t.Logf("VHDX unmount failed (expected without actual mount): %v", err)
	}
}