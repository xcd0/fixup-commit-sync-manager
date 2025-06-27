//go:build windows
// +build windows

package vhdx

import (
	"fmt"

	"github.com/Microsoft/go-winio/vhd"
)

// VirtualDisk はWindows環境でのみ使用するgo-winio VHD handle。
type VirtualDisk = vhd.VirtualDisk

// createVHDXWithGoWinio はgo-winioを使用してVHDXを作成する（Windows専用）。
func (v *VHDXManager) createVHDXWithGoWinio() error {
	sizeInBytes := v.parseSizeToBytes()
	
	// VHDX作成のパラメータを設定。
	params := &vhd.CreateVirtualDiskParameters{
		Version: 1,
		UniqueId: nil, // 自動生成
		MaximumSize: sizeInBytes,
		BlockSizeInBytes: 0, // デフォルト値を使用
		SectorSizeInBytes: 0, // デフォルト値を使用
		ParentPath: "",
		SourcePath: "",
	}

	// VHDXファイルを作成。
	handle, err := vhd.CreateVirtualDisk(v.VHDXPath, vhd.VirtualDiskAccessNone, vhd.CreateVirtualDiskFlagNone, params)
	if err != nil {
		return fmt.Errorf("failed to create VHDX with go-winio: %w", err)
	}
	
	// ハンドルを保存してクローズ。
	v.handle = handle
	defer v.handle.Close()

	// VHDX作成後、フォーマットとマウントを実行。
	return v.initializeAndFormatVHDX()
}

// mountVHDXWithGoWinio はgo-winioを使用してVHDXをマウントする（Windows専用）。
func (v *VHDXManager) mountVHDXWithGoWinio() error {
	// VHDXをオープン。
	handle, err := vhd.OpenVirtualDisk(v.VHDXPath, vhd.VirtualDiskAccessNone, vhd.OpenVirtualDiskFlagNone)
	if err != nil {
		return fmt.Errorf("failed to open VHDX with go-winio: %w", err)
	}
	
	v.handle = handle
	
	// VHDXをアタッチ。
	err = handle.Attach(vhd.AttachVirtualDiskFlagNone, nil)
	if err != nil {
		handle.Close()
		return fmt.Errorf("failed to attach VHDX with go-winio: %w", err)
	}

	// PowerShellでドライブレター割り当て（go-winioでは直接対応していない機能）。
	return v.assignDriveLetter()
}

// unmountVHDXWithGoWinio はgo-winioを使用してVHDXをアンマウントする（Windows専用）。
func (v *VHDXManager) unmountVHDXWithGoWinio() error {
	// 既存のハンドルがあるかチェック（ゼロ値と比較）。
	zeroHandle := vhd.VirtualDisk{}
	if v.handle == zeroHandle {
		handle, err := vhd.OpenVirtualDisk(v.VHDXPath, vhd.VirtualDiskAccessNone, vhd.OpenVirtualDiskFlagNone)
		if err != nil {
			return fmt.Errorf("failed to open VHDX for unmount: %w", err)
		}
		v.handle = handle
	}

	// VHDXをデタッチ。
	err := v.handle.Detach()
	if err != nil {
		return fmt.Errorf("failed to detach VHDX with go-winio: %w", err)
	}

	// ハンドルをクローズ。
	v.handle.Close()
	v.handle = zeroHandle

	return nil
}

// createSnapshotWithGoWinio はgo-winioを使用してスナップショットを作成する（Windows専用）。
func (v *VHDXManager) createSnapshotWithGoWinio(snapshotPath string) error {
	// 差分VHDXを作成するパラメータを設定。
	params := &vhd.CreateVirtualDiskParameters{
		Version: 1,
		UniqueId: nil, // 自動生成
		MaximumSize: 0, // 親VHDXから継承
		BlockSizeInBytes: 0, // デフォルト値を使用
		SectorSizeInBytes: 0, // デフォルト値を使用
		ParentPath: v.VHDXPath, // 親VHDXパス
		SourcePath: "",
	}

	// 差分VHDXファイルを作成。
	handle, err := vhd.CreateVirtualDisk(snapshotPath, vhd.VirtualDiskAccessNone, vhd.CreateVirtualDiskFlagNone, params)
	if err != nil {
		return fmt.Errorf("failed to create snapshot with go-winio: %w", err)
	}
	
	defer handle.Close()
	return nil
}

// hasGoWinioSupport はgo-winioサポートがあるかどうかを返す。
func (v *VHDXManager) hasGoWinioSupport() bool {
	return true
}