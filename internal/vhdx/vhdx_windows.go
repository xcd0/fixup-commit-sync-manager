//go:build windows
// +build windows

package vhdx

import (
	"syscall"
)

// VirtualDisk はWindows環境でのみ使用するsyscall.Handle。
type VirtualDisk = syscall.Handle

// createVHDXWithGoWinio はPowerShellを使用してVHDXを作成する（Windows専用）。
func (v *VHDXManager) createVHDXWithGoWinio() error {
	// go-winioの代わりにPowerShellでVHDX作成を実行。
	return v.createVHDXWithPowerShell()
}

// mountVHDXWithGoWinio はPowerShellを使用してVHDXをマウントする（Windows専用）。
func (v *VHDXManager) mountVHDXWithGoWinio() error {
	// PowerShellでVHDXをマウント。
	return v.mountVHDXWithPowerShell()
}

// unmountVHDXWithGoWinio はPowerShellを使用してVHDXをアンマウントする（Windows専用）。
func (v *VHDXManager) unmountVHDXWithGoWinio() error {
	// PowerShellでVHDXをアンマウント。
	return v.unmountVHDXWithPowerShell()
}

// createSnapshotWithGoWinio はPowerShellを使用してスナップショットを作成する（Windows専用）。
func (v *VHDXManager) createSnapshotWithGoWinio(snapshotPath string) error {
	// PowerShellでスナップショットを作成。
	return v.createSnapshotWithPowerShell(snapshotPath)
}

// hasGoWinioSupport はgo-winioサポートがあるかどうかを返す。
func (v *VHDXManager) hasGoWinioSupport() bool {
	return true
}