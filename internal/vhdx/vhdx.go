package vhdx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type VHDXManager struct {
	VHDXPath   string
	MountPoint string
	Size       string
	Encrypted  bool
}

func NewVHDXManager(vhdxPath, mountPoint, size string, encrypted bool) *VHDXManager {
	return &VHDXManager{
		VHDXPath:   vhdxPath,
		MountPoint: mountPoint,
		Size:       size,
		Encrypted:  encrypted,
	}
}

func (v *VHDXManager) CreateVHDX() error {
	if _, err := os.Stat(v.VHDXPath); err == nil {
		return fmt.Errorf("VHDX file already exists: %s", v.VHDXPath)
	}

	vhdxDir := filepath.Dir(v.VHDXPath)
	if err := os.MkdirAll(vhdxDir, 0755); err != nil {
		return fmt.Errorf("failed to create VHDX directory: %w", err)
	}

	diskpartScript := v.generateDiskpartScript()
	scriptPath := filepath.Join(vhdxDir, "create_vhdx.txt")

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

func (v *VHDXManager) MountVHDX() error {
	if v.isMounted() {
		return fmt.Errorf("VHDX is already mounted at %s", v.MountPoint)
	}

	diskpartScript := fmt.Sprintf(`select vdisk file="%s"
attach vdisk
assign letter=%s
exit
`, v.VHDXPath, strings.TrimSuffix(v.MountPoint, ":"))

	return v.executeDiskpartScript(diskpartScript)
}

func (v *VHDXManager) UnmountVHDX() error {
	if !v.isMounted() {
		return fmt.Errorf("VHDX is not mounted at %s", v.MountPoint)
	}

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

	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`
		$vhd = "%s"
		$snapshot = "%s"
		New-VHD -Path $snapshot -ParentPath $vhd -Differencing
	`, v.VHDXPath, snapshotPath))

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
