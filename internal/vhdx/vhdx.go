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
	
	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to convert WSL path to Windows: %w", err)
	}
	
	// VHDX作成後の初期化処理をPowerShellで実行。
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
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
	
	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to convert WSL path to Windows: %w", err)
	}
	
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
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

	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(absPath)
	if err != nil {
		return fmt.Errorf("failed to convert WSL path to Windows: %w", err)
	}

	// ディレクトリが存在することを確認。
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create VHDX directory: %w", err)
	}

	// PowerShellスクリプトでVHDX作成。
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
		fmt.Sprintf("$sizeBytes = %d", sizeInBytes),
		"",
		"# 既存ファイルがある場合は削除",
		"if (Test-Path $vhdxPath) {",
		"    Remove-Item $vhdxPath -Force",
		"    Write-Output \"Removed existing VHDX file: $vhdxPath\"",
		"}",
		"",
		"# ディレクトリが存在しない場合は作成",
		"$dir = Split-Path $vhdxPath -Parent",
		"if (!(Test-Path $dir)) {",
		"    New-Item -ItemType Directory -Path $dir -Force | Out-Null",
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
	v.VHDXPath = windowsPath

	// WSL環境またはテンポラリディレクトリではVHDX作成のみ行い、初期化をスキップ。
	if strings.Contains(windowsPath, "wsl.localhost") || 
	   strings.Contains(windowsPath, "Temp\\fixup-commit-sync-manager") ||
	   strings.Contains(windowsPath, "AppData\\Local\\Temp") {
		fmt.Printf("VHDX created successfully: %s\n", windowsPath)
		return nil
	}
	
	// VHDX作成後、フォーマットとマウントを実行。
	return v.initializeAndFormatVHDX()
}

// mountVHDXWithPowerShell はPowerShellを使用してVHDXをマウントする。
func (v *VHDXManager) mountVHDXWithPowerShell() error {
	// Tドライブのみ使用可能に制限。
	if !strings.HasPrefix(v.MountPoint, "T") {
		return fmt.Errorf("only T: drive is allowed for VHDX mounting, got: %s", v.MountPoint)
	}
	
	targetLetter := strings.TrimSuffix(v.MountPoint, ":")
	
	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to convert WSL path to Windows: %w", err)
	}
	
	// まず既存のマウントを確認してアンマウント。
	if err := v.ensureVHDXUnmounted(windowsPath); err != nil {
		return fmt.Errorf("failed to ensure VHDX unmounted: %w", err)
	}
	
	// VHDXマウントを実行。
	if err := v.mountVHDXOnly(windowsPath, targetLetter); err != nil {
		// マウント失敗時は強制アンマウントしてから再試行。
		v.forceUnmountVHDX(windowsPath)
		
		// 再試行。
		if retryErr := v.mountVHDXOnly(windowsPath, targetLetter); retryErr != nil {
			return fmt.Errorf("failed to mount VHDX after retry: %w (original error: %v)", retryErr, err)
		}
	}
	
	// マウント後、即座にアンマウント。
	return v.unmountVHDXWithPowerShell()
}

// unmountVHDXWithPowerShell はPowerShellを使用してVHDXをアンマウントする。
func (v *VHDXManager) unmountVHDXWithPowerShell() error {
	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(v.VHDXPath)
	if err != nil {
		return fmt.Errorf("failed to convert WSL path to Windows: %w", err)
	}
	
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
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

// convertWSLPathToWindows はWSLパスをWindows形式に変換する。
func (v *VHDXManager) convertWSLPathToWindows(path string) (string, error) {
	// 既にWindows形式の場合はそのまま返す。
	if !strings.Contains(path, "wsl.localhost") && !strings.HasPrefix(path, "/") {
		return path, nil
	}
	
	// WSL内のUnixパスまたはwsl.localhostパスの場合。
	if strings.HasPrefix(path, "/") || strings.Contains(path, "wsl.localhost") {
		// Windows側のテンポラリディレクトリを使用。
		return v.manualWSLPathConversion(path)
	}
	
	return path, nil
}

// manualWSLPathConversion は手動でWSLパスをWindows形式に変換する。
func (v *VHDXManager) manualWSLPathConversion(path string) (string, error) {
	// ファイル名のみを抽出。
	fileName := filepath.Base(path)
	
	// Windows側のテンポラリディレクトリにVHDXを作成。
	windowsTempPath := os.Getenv("TEMP")
	if windowsTempPath == "" {
		// WSLから見えるWindows Cドライブのパスを使用。
		windowsTempPath = "/mnt/c/Temp"
		
		// PowerShell用にWindows形式に変換。
		windowsPath := "C:\\Temp\\fixup-commit-sync-manager\\" + fileName
		return windowsPath, nil
	}
	
	// Windows環境変数TEMPが取得できた場合はそれを使用。
	windowsPath := filepath.Join(windowsTempPath, "fixup-commit-sync-manager", fileName)
	windowsPath = strings.ReplaceAll(windowsPath, "/", "\\")
	
	return windowsPath, nil
}

// ensureVHDXUnmounted は既存のVHDXマウントを確認してアンマウントする。
func (v *VHDXManager) ensureVHDXUnmounted(windowsPath string) error {
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
		"",
		"# 既存のVHDマウントを確認",
		"try {",
		"    $vhd = Get-VHD -Path $vhdxPath -ErrorAction SilentlyContinue",
		"    if ($vhd -and $vhd.Attached) {",
		"        Write-Output \"VHDX is currently mounted, dismounting...\"",
		"        Dismount-VHD -Path $vhdxPath -ErrorAction SilentlyContinue",
		"        Start-Sleep -Seconds 2",
		"    }",
		"} catch {",
		"    Write-Output \"No existing VHDX mount found or error checking: $($_.Exception.Message)\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")
	
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// アンマウント前のチェックでエラーが出ても継続。
		fmt.Printf("Warning during VHDX unmount check: %s\n", string(output))
	}
	
	return nil
}

// mountVHDXOnly はVHDXマウントのみを実行する。
func (v *VHDXManager) mountVHDXOnly(windowsPath, targetLetter string) error {
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
		fmt.Sprintf("$targetLetter = \"%s\"", targetLetter),
		"",
		"# VHDXをマウント",
		"try {",
		"    $disk = Mount-VHD -Path $vhdxPath -PassThru | Get-Disk",
		"    ",
		"    if ($disk) {",
		"        # パーティションを取得",
		"        $partitions = Get-Partition -DiskNumber $disk.Number -ErrorAction SilentlyContinue",
		"        $partition = $partitions | Where-Object Type -eq 'Basic' | Select-Object -First 1",
		"        ",
		"        if (!$partition) {",
		"            # パーティションが存在しない場合は初期化",
		"            Write-Output \"No partition found, initializing disk...\"",
		"            $disk | Initialize-Disk -PartitionStyle GPT -PassThru |",
		"            New-Partition -AssignDriveLetter -UseMaximumSize |",
		"            Format-Volume -FileSystem NTFS -NewFileSystemLabel \"VHDX\" -Confirm:$false",
		"            ",
		"            # 新しく作成されたパーティションを取得",
		"            $partition = Get-Partition -DiskNumber $disk.Number | Where-Object Type -eq 'Basic' | Select-Object -First 1",
		"        }",
		"        ",
		"        if ($partition) {",
		"            # 指定されたドライブレターに変更",
		"            if ($partition.DriveLetter -ne $targetLetter) {",
		"                $partition | Set-Partition -NewDriveLetter $targetLetter",
		"            }",
		"            Write-Output \"VHDX mounted successfully at $($targetLetter):\"",
		"        } else {",
		"            throw \"Failed to create or find partition on the disk\"",
		"        }",
		"    } else {",
		"        throw \"Failed to get disk information after mounting VHDX\"",
		"    }",
		"} catch {",
		"    throw \"VHDX mount failed: $($_.Exception.Message)\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount VHDX: %w, output: %s", err, string(output))
	}

	return nil
}

// forceUnmountVHDX は強制的にVHDXをアンマウントする。
func (v *VHDXManager) forceUnmountVHDX(windowsPath string) {
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
		"",
		"# 強制アンマウント",
		"try {",
		"    Dismount-VHD -Path $vhdxPath -ErrorAction SilentlyContinue",
		"    Start-Sleep -Seconds 1",
		"    Write-Output \"Force unmount completed\"",
		"} catch {",
		"    Write-Output \"Force unmount failed or not needed: $($_.Exception.Message)\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")
	
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Warning during force unmount: %s\n", string(output))
	}
}

// isVHDXMountedWithPowerShell はPowerShellでVHDXのマウント状態を確認する。
func (v *VHDXManager) isVHDXMountedWithPowerShell() bool {
	// WSLパスをWindows形式に変換。
	windowsPath, err := v.convertWSLPathToWindows(v.VHDXPath)
	if err != nil {
		return false
	}
	
	scriptLines := []string{
		fmt.Sprintf("$vhdxPath = \"%s\"", windowsPath),
		"",
		"# VHDマウント状態を確認",
		"try {",
		"    $vhd = Get-VHD -Path $vhdxPath -ErrorAction SilentlyContinue",
		"    if ($vhd -and $vhd.Attached) {",
		"        Write-Output \"true\"",
		"    } else {",
		"        Write-Output \"false\"",
		"    }",
		"} catch {",
		"    Write-Output \"false\"",
		"}",
	}
	script := strings.Join(scriptLines, "\n")
	
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.TrimSpace(string(output)) == "true"
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
	// Windows環境では実際のVHDマウント状態をPowerShellで確認。
	if runtime.GOOS == "windows" {
		return v.isVHDXMountedWithPowerShell()
	}
	
	// 他の環境ではマウントポイントディレクトリの存在確認。
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
