package vhdx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type VHDXManager struct {
	VHDXPath   string
	MountPoint string
	Size       string
	Encrypted  bool
	handle     VirtualDisk // プラットフォーム固有のVHD handle
}

func NewVHDXManager(vhdxPath, mountPoint, size string, encrypted bool) *VHDXManager {
	return &VHDXManager{
		VHDXPath:   vhdxPath,
		MountPoint: mountPoint,
		Size:       size,
		Encrypted:  encrypted,
	}
}

// NewManager は既存のコンストラクタとの互換性を保つ。
func NewManager(vhdxPath, mountPoint string) *VHDXManager {
	return &VHDXManager{
		VHDXPath:   vhdxPath,
		MountPoint: mountPoint,
		Size:       "10GB",
		Encrypted:  false,
	}
}

// Create はVHDXファイルを作成する。
func (v *VHDXManager) Create(size string, encrypted bool) error {
	if size != "" {
		v.Size = size
	}
	v.Encrypted = encrypted
	return v.CreateVHDX()
}

func (v *VHDXManager) CreateVHDX() error {
	if _, err := os.Stat(v.VHDXPath); err == nil {
		return fmt.Errorf("VHDX file already exists: %s", v.VHDXPath)
	}

	vhdxDir := filepath.Dir(v.VHDXPath)
	if err := os.MkdirAll(vhdxDir, 0755); err != nil {
		return fmt.Errorf("failed to create VHDX directory: %w", err)
	}

	// Windows環境でgo-winioを使用してVHDX作成。
	if runtime.GOOS == "windows" {
		return v.createVHDXWithGoWinio()
	}

	// Windows以外の環境では従来のdiskpart方式（互換性のため）。
	return v.createVHDXWithDiskpart()
}

// createVHDXWithGoWinio の実装はプラットフォーム固有ファイルに移動。

// createVHDXWithDiskpart は従来のdiskpart方式でVHDXを作成する。
func (v *VHDXManager) createVHDXWithDiskpart() error {
	diskpartScript := v.generateDiskpartScript()
	scriptPath := filepath.Join(filepath.Dir(v.VHDXPath), "create_vhdx.txt")

	if err := os.WriteFile(scriptPath, []byte(diskpartScript), 0644); err != nil {
		return fmt.Errorf("failed to write diskpart script: %w", err)
	}
	defer os.Remove(scriptPath)

	cmd := exec.Command("diskpart", "/s", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskpart failed: %w, output: %s", err, string(output))
	}

	return nil
}

// initializeAndFormatVHDX はVHDXを初期化してフォーマットする。
func (v *VHDXManager) initializeAndFormatVHDX() error {
	targetLetter := strings.TrimSuffix(v.MountPoint, ":")
	
	// VHDX作成後の初期化処理をPowerShellで実行。
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", v.VHDXPath),
		fmt.Sprintf("$targetLetter = \"%s\"", targetLetter),
		"",
		"# VHDXをアタッチ",
		"$disk = Mount-VHD -Path $vhdxPath -PassThru | Get-Disk",
		"",
		"# パーティション初期化",
		"$disk | Initialize-Disk -PartitionStyle GPT -PassThru |",
		"New-Partition -AssignDriveLetter -UseMaximumSize |",
		"Format-Volume -FileSystem NTFS -NewFileSystemLabel \"VHDX\" -Confirm:$false",
		"",
		"# 指定されたドライブレターに変更",
		"$partition = Get-Partition -DiskNumber $disk.Number | Where-Object Type -eq 'Basic'",
		"if ($partition) {",
		"    $partition | Set-Partition -NewDriveLetter $targetLetter",
		"}",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize VHDX: %w, output: %s", err, string(output))
	}

	return nil
}

// Mount はVHDXをマウントする。
func (v *VHDXManager) Mount() error {
	return v.MountVHDX()
}

func (v *VHDXManager) MountVHDX() error {
	if v.isMounted() {
		return fmt.Errorf("VHDX is already mounted at %s", v.MountPoint)
	}

	// Windows環境でgo-winioを使用してマウント。
	if runtime.GOOS == "windows" {
		return v.mountVHDXWithGoWinio()
	}

	// Windows以外の環境では従来のdiskpart方式（互換性のため）。
	return v.mountVHDXWithDiskpart()
}

// mountVHDXWithGoWinio の実装はプラットフォーム固有ファイルに移動。

// mountVHDXWithDiskpart は従来のdiskpart方式でマウントする。
func (v *VHDXManager) mountVHDXWithDiskpart() error {
	diskpartScript := fmt.Sprintf(`select vdisk file="%s"
attach vdisk
assign letter=%s
exit
`, v.VHDXPath, strings.TrimSuffix(v.MountPoint, ":"))

	return v.executeDiskpartScript(diskpartScript)
}

// assignDriveLetter はVHDXに指定されたドライブレターを割り当てる。
func (v *VHDXManager) assignDriveLetter() error {
	targetLetter := strings.TrimSuffix(v.MountPoint, ":")
	
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", v.VHDXPath),
		fmt.Sprintf("$targetLetter = \"%s\"", targetLetter),
		"",
		"# VHDXに関連付けられたディスクを取得",
		"$disk = Get-VHD -Path $vhdxPath | Get-Disk",
		"if ($disk) {",
		"    $partition = Get-Partition -DiskNumber $disk.Number | Where-Object Type -eq 'Basic'",
		"    if ($partition -and $partition.DriveLetter -ne $targetLetter) {",
		"        $partition | Set-Partition -NewDriveLetter $targetLetter",
		"    }",
		"}",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to assign drive letter: %w, output: %s", err, string(output))
	}

	return nil
}

func (v *VHDXManager) UnmountVHDX() error {
	if !v.isMounted() {
		return fmt.Errorf("VHDX is not mounted at %s", v.MountPoint)
	}

	// Windows環境でgo-winioを使用してアンマウント。
	if runtime.GOOS == "windows" {
		return v.unmountVHDXWithGoWinio()
	}

	// Windows以外の環境では従来のdiskpart方式（互換性のため）。
	return v.unmountVHDXWithDiskpart()
}

// unmountVHDXWithGoWinio の実装はプラットフォーム固有ファイルに移動。

// unmountVHDXWithDiskpart は従来のdiskpart方式でアンマウントする。
func (v *VHDXManager) unmountVHDXWithDiskpart() error {
	diskpartScript := fmt.Sprintf(`select vdisk file="%s"
detach vdisk
exit
`, v.VHDXPath)

	return v.executeDiskpartScript(diskpartScript)
}

func (v *VHDXManager) CreateSnapshot(name string) error {
	if name == "" {
		name = fmt.Sprintf("snapshot_%d", time.Now().Unix())
	}

	snapshotPath := v.getSnapshotPath(name)

	// スナップショットディレクトリを作成。
	snapshotDir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Windows環境でgo-winioを使用してスナップショット作成。
	if runtime.GOOS == "windows" {
		return v.createSnapshotWithGoWinio(snapshotPath)
	}

	// Windows以外の環境では従来のPowerShell方式（互換性のため）。
	return v.createSnapshotWithPowerShell(snapshotPath)
}

// createVHDXWithPowerShell はPowerShellを使用してVHDXを作成する。
func (v *VHDXManager) createVHDXWithPowerShell() error {
	sizeInBytes := v.parseSizeToBytes()
	sizeInGB := sizeInBytes / (1024 * 1024 * 1024)
	if sizeInGB == 0 {
		sizeInGB = 1 // 最小1GB
	}

	// VHDXファイルパスを絶対パスに変換。
	absPath, err := filepath.Abs(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to convert VHDX path to absolute: %w", err)
	}

	// ディレクトリが存在することを確認。
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create VHDX directory: %w", err)
	}

	// PowerShellスクリプトでVHDX作成。
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", absPath),
		fmt.Sprintf("$sizeBytes = %d", sizeInBytes),
		"",
		"# 既存ファイルチェック",
		"if (Test-Path $vhdxPath) {",
		"    throw \"VHDX file already exists: $vhdxPath\"",
		"}",
		"",
		"# VHDXを作成",
		"New-VHD -Path $vhdxPath -SizeBytes $sizeBytes -Dynamic",
		"",
		"Write-Output \"VHDX created successfully: $vhdxPath ($sizeBytes bytes)\"",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create VHDX with PowerShell (path=%s, size=%d bytes): %w, output: %s", 
			absPath, sizeInBytes, err, string(output))
	}

	// VHDX作成成功後、パスを更新。
	v.VHDXPath = absPath

	// WSL環境ではVHDX作成のみ行い、初期化は後で実行。
	if strings.Contains(absPath, "wsl.localhost") {
		fmt.Printf("VHDX created successfully (initialization skipped in WSL): %s\n", absPath)
		return nil
	}
	
	// VHDX作成後、フォーマットとマウントを実行。
	return v.initializeAndFormatVHDX()
}

// mountVHDXWithPowerShell はPowerShellを使用してVHDXをマウントする。
func (v *VHDXManager) mountVHDXWithPowerShell() error {
	targetLetter := strings.TrimSuffix(v.MountPoint, ":")
	
	// PowerShellスクリプトを個別に構築してGoの変数展開問題を回避。
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", v.VHDXPath),
		fmt.Sprintf("$targetLetter = \"%s\"", targetLetter),
		"",
		"# VHDXをマウント",
		"$disk = Mount-VHD -Path $vhdxPath -PassThru | Get-Disk",
		"",
		"if ($disk) {",
		"    # パーティションを取得",
		"    $partition = Get-Partition -DiskNumber $disk.Number | Where-Object Type -eq 'Basic' | Select-Object -First 1",
		"    ",
		"    if ($partition) {",
		"        # 指定されたドライブレターに変更",
		"        if ($partition.DriveLetter -ne $targetLetter) {",
		"            $partition | Set-Partition -NewDriveLetter $targetLetter",
		"        }",
		"        Write-Output \"VHDX mounted successfully at $targetLetter:\"",
		"    } else {",
		"        Write-Warning \"No basic partition found on the disk\"",
		"    }",
		"} else {",
		"    throw \"Failed to get disk information after mounting VHDX\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount VHDX with PowerShell: %w, output: %s", err, string(output))
	}

	return nil
}

// unmountVHDXWithPowerShell はPowerShellを使用してVHDXをアンマウントする。
func (v *VHDXManager) unmountVHDXWithPowerShell() error {
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", v.VHDXPath),
		"",
		"# VHDXをアンマウント",
		"try {",
		"    Dismount-VHD -Path $vhdxPath",
		"    Write-Output \"VHDX unmounted successfully: $vhdxPath\"",
		"} catch {",
		"    throw \"Failed to unmount VHDX: $($_.Exception.Message)\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unmount VHDX with PowerShell: %w, output: %s", err, string(output))
	}

	// ハンドルをリセット。
	v.handle = 0

	return nil
}

// createSnapshotWithGoWinio の実装はプラットフォーム固有ファイルに移動。

// createSnapshotWithPowerShell は従来のPowerShell方式でスナップショットを作成する。
func (v *VHDXManager) createSnapshotWithPowerShell(snapshotPath string) error {
	scriptLines := []string{
		fmt.Sprintf("$vhd = \"%s\"", v.VHDXPath),
		fmt.Sprintf("$snapshot = \"%s\"", snapshotPath),
		"New-VHD -Path $snapshot -ParentPath $vhd -Differencing",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w, output: %s", err, string(output))
	}

	return nil
}

func (v *VHDXManager) ListSnapshots() ([]string, error) {
	snapshotDir := filepath.Join(filepath.Dir(v.VHDXPath), "snapshots")

	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	var snapshots []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".vhdx") {
			snapshots = append(snapshots, strings.TrimSuffix(entry.Name(), ".vhdx"))
		}
	}

	return snapshots, nil
}

func (v *VHDXManager) RollbackToSnapshot(name string) error {
	snapshotPath := v.getSnapshotPath(name)

	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot not found: %s", name)
	}

	if v.isMounted() {
		if err := v.UnmountVHDX(); err != nil {
			return fmt.Errorf("failed to unmount VHDX before rollback: %w", err)
		}
	}

	backupPath := v.VHDXPath + ".backup"
	if err := os.Rename(v.VHDXPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current VHDX: %w", err)
	}

	if err := copyFile(snapshotPath, v.VHDXPath); err != nil {
		os.Rename(backupPath, v.VHDXPath)
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	os.Remove(backupPath)
	return nil
}

func (v *VHDXManager) isMounted() bool {
	_, err := os.Stat(v.MountPoint)
	return err == nil
}

func (v *VHDXManager) generateDiskpartScript() string {
	sizeInMB := v.parseSizeToMB()

	script := fmt.Sprintf(`create vdisk file="%s" maximum=%d type=expandable
select vdisk file="%s"
attach vdisk
create partition primary
format fs=ntfs quick
assign letter=%s
exit
`, v.VHDXPath, sizeInMB, v.VHDXPath, strings.TrimSuffix(v.MountPoint, ":"))

	return script
}

func (v *VHDXManager) parseSizeToMB() int {
	size := strings.ToUpper(v.Size)

	if strings.HasSuffix(size, "GB") {
		sizeStr := strings.TrimSuffix(size, "GB")
		if gb := parseInt(sizeStr); gb > 0 {
			return gb * 1024
		}
	} else if strings.HasSuffix(size, "MB") {
		sizeStr := strings.TrimSuffix(size, "MB")
		if mb := parseInt(sizeStr); mb > 0 {
			return mb
		}
	}

	return 10240
}

// parseSizeToBytes はサイズ文字列をバイト数に変換する（go-winio用）。
func (v *VHDXManager) parseSizeToBytes() uint64 {
	size := strings.ToUpper(v.Size)

	if strings.HasSuffix(size, "GB") {
		sizeStr := strings.TrimSuffix(size, "GB")
		if gb := parseInt(sizeStr); gb > 0 {
			return uint64(gb) * 1024 * 1024 * 1024
		}
	} else if strings.HasSuffix(size, "MB") {
		sizeStr := strings.TrimSuffix(size, "MB")
		if mb := parseInt(sizeStr); mb > 0 {
			return uint64(mb) * 1024 * 1024
		}
	} else if strings.HasSuffix(size, "TB") {
		sizeStr := strings.TrimSuffix(size, "TB")
		if tb := parseInt(sizeStr); tb > 0 {
			return uint64(tb) * 1024 * 1024 * 1024 * 1024
		}
	}

	// デフォルト: 10GB
	return 10 * 1024 * 1024 * 1024
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}

func (v *VHDXManager) executeDiskpartScript(script string) error {
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, fmt.Sprintf("diskpart_%d.txt", time.Now().UnixNano()))

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write diskpart script: %w", err)
	}
	defer os.Remove(scriptPath)

	cmd := exec.Command("diskpart", "/s", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskpart failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (v *VHDXManager) getSnapshotPath(name string) string {
	snapshotDir := filepath.Join(filepath.Dir(v.VHDXPath), "snapshots")
	return filepath.Join(snapshotDir, name+".vhdx")
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}
