//go:build windows
// +build windows

package vhdx

import (
	"fmt"
	"syscall"

	"github.com/Microsoft/go-winio/vhd"
)

// VirtualDisk はWindows環境でのみ使用するsyscall.Handle。
type VirtualDisk = syscall.Handle

// createVHDXWithGoWinio はgo-winioを使用してVHDXを作成する（Windows専用）。
func (v *VHDXManager) createVHDXWithGoWinio() error {
	sizeInBytes := v.parseSizeToBytes()
	
	// 簡易VHDX作成（go-winioのヘルパー関数を使用）。
	sizeInGB := uint32(sizeInBytes / (1024 * 1024 * 1024))
	if sizeInGB == 0 {
		sizeInGB = 1 // 最小1GB
	}
	
	err := vhd.CreateVhdx(v.VHDXPath, sizeInGB, 0) // 0 = デフォルトブロックサイズ
	if err != nil {
		return fmt.Errorf("failed to create VHDX with go-winio: %w", err)
	}

	// VHDX作成後、フォーマットとマウントを実行。
	return v.initializeAndFormatVHDX()
}

// mountVHDXWithGoWinio はgo-winioを使用してVHDXをマウントする（Windows専用）。
func (v *VHDXManager) mountVHDXWithGoWinio() error {
	// 簡易VHDアタッチ（go-winioのヘルパー関数を使用）。
	err := vhd.AttachVhd(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to attach VHDX with go-winio: %w", err)
	}

	// PowerShellでドライブレター割り当て（go-winioでは直接対応していない機能）。
	return v.assignDriveLetter()
}

// unmountVHDXWithGoWinio はgo-winioを使用してVHDXをアンマウントする（Windows専用）。
func (v *VHDXManager) unmountVHDXWithGoWinio() error {
	// 簡易VHDデタッチ（go-winioのヘルパー関数を使用）。
	err := vhd.DetachVhd(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to detach VHDX with go-winio: %w", err)
	}

	// ハンドルをリセット。
	v.handle = 0

	return nil
}

// createSnapshotWithGoWinio はgo-winioを使用してスナップショットを作成する（Windows専用）。
func (v *VHDXManager) createSnapshotWithGoWinio(snapshotPath string) error {
	// 差分VHDを作成（go-winioのヘルパー関数を使用）。
	err := vhd.CreateDiffVhd(snapshotPath, v.VHDXPath, 0) // 0 = デフォルトブロックサイズ
	if err != nil {
		return fmt.Errorf("failed to create snapshot with go-winio: %w", err)
	}
	
	return nil
}

// hasGoWinioSupport はgo-winioサポートがあるかどうかを返す。
func (v *VHDXManager) hasGoWinioSupport() bool {
	return true
}