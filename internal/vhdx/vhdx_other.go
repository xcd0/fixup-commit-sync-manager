//go:build !windows
// +build !windows

package vhdx

import (
	"fmt"
)

// VirtualDisk は非Windows環境でのプレースホルダー型。
type VirtualDisk struct{}

// Close はプレースホルダーメソッド。
func (vd VirtualDisk) Close() error {
	return nil
}

// createVHDXWithGoWinio は非Windows環境では使用できない。
func (v *VHDXManager) createVHDXWithGoWinio() error {
	return fmt.Errorf("go-winio VHDX creation is only supported on Windows")
}

// mountVHDXWithGoWinio は非Windows環境では使用できない。
func (v *VHDXManager) mountVHDXWithGoWinio() error {
	return fmt.Errorf("go-winio VHDX mounting is only supported on Windows")
}

// unmountVHDXWithGoWinio は非Windows環境では使用できない。
func (v *VHDXManager) unmountVHDXWithGoWinio() error {
	return fmt.Errorf("go-winio VHDX unmounting is only supported on Windows")
}

// createSnapshotWithGoWinio は非Windows環境では使用できない。
func (v *VHDXManager) createSnapshotWithGoWinio(snapshotPath string) error {
	return fmt.Errorf("go-winio VHDX snapshot creation is only supported on Windows")
}

// hasGoWinioSupport は非Windows環境ではfalseを返す。
func (v *VHDXManager) hasGoWinioSupport() bool {
	return false
}