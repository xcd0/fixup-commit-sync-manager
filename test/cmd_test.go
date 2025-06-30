package test

import (
	"strings"
	"testing"

	"fixup-commit-sync-manager/cmd"
)

// TestCmd1_InitConfigConstructor はinit-configコマンドコンストラクタの基本テスト（TDD Step 1）
func TestCmd1_InitConfigConstructor(t *testing.T) {
	t.Log("Cmd Test 1: Init config constructor")
	
	// init-configコマンドの作成
	initConfigCmd := cmd.NewInitConfigCmd()
	if initConfigCmd == nil {
		t.Fatal("Failed to create init-config command")
	}
	
	// 基本的なコマンド情報確認
	if initConfigCmd.Use != "init-config" {
		t.Errorf("Expected Use 'init-config', got '%s'", initConfigCmd.Use)
	}
	
	if initConfigCmd.Short == "" {
		t.Error("Init-config command should have Short description")
	}
	
	if initConfigCmd.RunE == nil {
		t.Error("Init-config command should have RunE function")
	}
	
	t.Log("✓ Init config constructor test OK")
}

// TestCmd2_ValidateConfigConstructor はvalidate-configコマンドコンストラクタのテスト（TDD Step 2）
func TestCmd2_ValidateConfigConstructor(t *testing.T) {
	t.Log("Cmd Test 2: Validate config constructor")
	
	// validate-configコマンドの作成
	validateCmd := cmd.NewValidateConfigCmd()
	if validateCmd == nil {
		t.Fatal("Failed to create validate-config command")
	}
	
	// コマンド情報確認
	if validateCmd.Use != "validate-config" {
		t.Errorf("Expected Use 'validate-config', got '%s'", validateCmd.Use)
	}
	
	if validateCmd.RunE == nil {
		t.Error("Validate-config command should have RunE function")
	}
	
	t.Log("✓ Validate config constructor test OK")
}

// TestCmd3_SyncConstructor はsyncコマンドコンストラクタのテスト（TDD Step 3）
func TestCmd3_SyncConstructor(t *testing.T) {
	t.Log("Cmd Test 3: Sync constructor")
	
	// syncコマンドの作成
	syncCmd := cmd.NewSyncCmd()
	if syncCmd == nil {
		t.Fatal("Failed to create sync command")
	}
	
	// コマンド情報確認
	if syncCmd.Use != "sync" {
		t.Errorf("Expected Use 'sync', got '%s'", syncCmd.Use)
	}
	
	if syncCmd.RunE == nil {
		t.Error("Sync command should have RunE function")
	}
	
	// フラグの確認
	continuousFlag := syncCmd.Flags().Lookup("continuous")
	if continuousFlag == nil {
		t.Error("Sync command should have --continuous flag")
	}
	
	t.Log("✓ Sync constructor test OK")
}

// TestCmd4_SyncCommand はsyncコマンドのテスト（TDD Step 4）
func TestCmd4_SyncCommand(t *testing.T) {
	t.Log("Cmd Test 4: Sync command")
	
	// syncコマンドの作成
	syncCmd := cmd.NewSyncCmd()
	if syncCmd == nil {
		t.Fatal("Failed to create sync command")
	}
	
	// コマンド情報確認
	if syncCmd.Use != "sync" {
		t.Errorf("Expected Use 'sync', got '%s'", syncCmd.Use)
	}
	
	if syncCmd.RunE == nil {
		t.Error("Sync command should have RunE function")
	}
	
	// フラグの確認
	continuousFlag := syncCmd.Flags().Lookup("continuous")
	if continuousFlag == nil {
		t.Error("Sync command should have --continuous flag")
	}
	
	t.Log("✓ Sync command test OK")
}

// TestCmd5_FixupCommand はfixupコマンドのテスト（TDD Step 5）
func TestCmd5_FixupCommand(t *testing.T) {
	t.Log("Cmd Test 5: Fixup command")
	
	// fixupコマンドの作成
	fixupCmd := cmd.NewFixupCmd()
	if fixupCmd == nil {
		t.Fatal("Failed to create fixup command")
	}
	
	// コマンド情報確認
	if fixupCmd.Use != "fixup" {
		t.Errorf("Expected Use 'fixup', got '%s'", fixupCmd.Use)
	}
	
	if fixupCmd.RunE == nil {
		t.Error("Fixup command should have RunE function")
	}
	
	// フラグの確認
	continuousFlag := fixupCmd.Flags().Lookup("continuous")
	if continuousFlag == nil {
		t.Error("Fixup command should have --continuous flag")
	}
	
	t.Log("✓ Fixup command test OK")
}

// TestCmd6_InitVHDXCommand はinit-vhdxコマンドのテスト（TDD Step 6）
func TestCmd6_InitVHDXCommand(t *testing.T) {
	t.Log("Cmd Test 6: Init VHDX command")
	
	// init-vhdxコマンドの作成
	initVHDXCmd := cmd.NewInitVHDXCmd()
	if initVHDXCmd == nil {
		t.Fatal("Failed to create init-vhdx command")
	}
	
	// コマンド情報確認
	if initVHDXCmd.Use != "init-vhdx" {
		t.Errorf("Expected Use 'init-vhdx', got '%s'", initVHDXCmd.Use)
	}
	
	if initVHDXCmd.RunE == nil {
		t.Error("Init-vhdx command should have RunE function")
	}
	
	t.Log("✓ Init VHDX command test OK")
}

// TestCmd7_MountVHDXCommand はmount-vhdxコマンドのテスト（TDD Step 7）
func TestCmd7_MountVHDXCommand(t *testing.T) {
	t.Log("Cmd Test 7: Mount VHDX command")
	
	// mount-vhdxコマンドの作成
	mountCmd := cmd.NewMountVHDXCmd()
	if mountCmd == nil {
		t.Fatal("Failed to create mount-vhdx command")
	}
	
	// コマンド情報確認
	if mountCmd.Use != "mount-vhdx" {
		t.Errorf("Expected Use 'mount-vhdx', got '%s'", mountCmd.Use)
	}
	
	if mountCmd.RunE == nil {
		t.Error("Mount-vhdx command should have RunE function")
	}
	
	t.Log("✓ Mount VHDX command test OK")
}

// TestCmd8_UnmountVHDXCommand はunmount-vhdxコマンドのテスト（TDD Step 8）
func TestCmd8_UnmountVHDXCommand(t *testing.T) {
	t.Log("Cmd Test 8: Unmount VHDX command")
	
	// unmount-vhdxコマンドの作成
	unmountCmd := cmd.NewUnmountVHDXCmd()
	if unmountCmd == nil {
		t.Fatal("Failed to create unmount-vhdx command")
	}
	
	// コマンド情報確認
	if unmountCmd.Use != "unmount-vhdx" {
		t.Errorf("Expected Use 'unmount-vhdx', got '%s'", unmountCmd.Use)
	}
	
	if unmountCmd.RunE == nil {
		t.Error("Unmount-vhdx command should have RunE function")
	}
	
	t.Log("✓ Unmount VHDX command test OK")
}

// TestCmd9_SnapshotVHDXCommand はsnapshot-vhdxコマンドのテスト（TDD Step 9）
func TestCmd9_SnapshotVHDXCommand(t *testing.T) {
	t.Log("Cmd Test 9: Snapshot VHDX command")
	
	// snapshot-vhdxコマンドの作成
	snapshotCmd := cmd.NewSnapshotVHDXCmd()
	if snapshotCmd == nil {
		t.Fatal("Failed to create snapshot-vhdx command")
	}
	
	// コマンド情報確認
	if snapshotCmd.Use != "snapshot-vhdx" {
		t.Errorf("Expected Use 'snapshot-vhdx', got '%s'", snapshotCmd.Use)
	}
	
	if snapshotCmd.RunE == nil {
		t.Error("Snapshot-vhdx command should have RunE function")
	}
	
	// サブコマンドの確認
	subcommands := []string{"create", "list", "rollback"}
	for _, subcmd := range subcommands {
		found := false
		for _, cmd := range snapshotCmd.Commands() {
			if cmd.Use == subcmd || strings.Contains(cmd.Use, subcmd) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Snapshot subcommand %s not found", subcmd)
		}
	}
	
	t.Log("✓ Snapshot VHDX command test OK")
}

// TestCmd10_GlobalFlags はグローバルフラグのテスト（TDD Step 10）
func TestCmd10_GlobalFlags(t *testing.T) {
	t.Log("Cmd Test 10: Global flags")
	
	rootCmd := cmd.NewRootCmd()
	
	// グローバルフラグの確認
	globalFlags := []string{"config", "dry-run", "verbose"}
	
	for _, flagName := range globalFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Global flag --%s not found", flagName)
		}
	}
	
	// デフォルト値の確認
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag != nil && configFlag.DefValue != "config.hjson" {
		t.Errorf("Config flag default value should be 'config.hjson', got '%s'", configFlag.DefValue)
	}
	
	t.Log("✓ Global flags test OK")
}