package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"fixup-commit-sync-manager/internal/vhdx"
)

// initCmd はFixupCommitSyncManagerの初期セットアップを行う。
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "FixupCommitSyncManagerの初期セットアップを実行",
	Long: `FixupCommitSyncManagerの初期セットアップを実行します。

このコマンドは以下の処理を行います:
1. 作業ディレクトリの作成
2. 実行ファイルのコピー
3. 設定ファイルの対話的生成

初回使用時に実行することを推奨します。
VHDXの初期化は別途 init-vhdx コマンドで実行してください。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	// initCmdはroot.goで登録されるため、ここでは追加しない
}

// runInit は初期セットアップを実行する。
func runInit() error {
	fmt.Println("=== FixupCommitSyncManager 初期セットアップ ===")
	fmt.Println("FixupCommitSyncManagerの作業環境を構築します。")
	fmt.Println()

	// 作業ディレクトリの決定。
	workDir, err := promptWorkingDirectory()
	if err != nil {
		return fmt.Errorf("作業ディレクトリの設定に失敗しました: %w", err)
	}

	// 作業ディレクトリの作成。
	if err := createWorkingDirectory(workDir); err != nil {
		return fmt.Errorf("作業ディレクトリの作成に失敗しました: %w", err)
	}

	// 実行ファイルのコピー。
	if err := copyExecutable(workDir); err != nil {
		return fmt.Errorf("実行ファイルのコピーに失敗しました: %w", err)
	}

	// 設定ファイルの生成。
	configPath := filepath.Join(workDir, "config.hjson")
	if err := generateInitialConfig(workDir, configPath); err != nil {
		return fmt.Errorf("設定ファイルの生成に失敗しました: %w", err)
	}

	// 完了メッセージ。
	fmt.Printf("\n    === セットアップ完了 ===\n")
	fmt.Printf("    作業ディレクトリ: %s\n", workDir)
	fmt.Printf("    実行ファイル: %s\n", filepath.Join(workDir, "fixup-commit-sync-manager.exe"))
	fmt.Printf("    設定ファイル: %s\n", configPath)
	fmt.Println()
	fmt.Println("    次の手順:")
	fmt.Println("    1. 設定ファイルを確認・編集してください")
	fmt.Printf("    2. VHDXを初期化: %s init-vhdx\n", filepath.Join(workDir, "fixup-commit-sync-manager.exe"))
	fmt.Printf("    3. VHDXをマウント: %s mount-vhdx\n", filepath.Join(workDir, "fixup-commit-sync-manager.exe"))
	fmt.Printf("    4. 同期開始: %s sync --continuous\n", filepath.Join(workDir, "fixup-commit-sync-manager.exe"))

	return nil
}

// promptWorkingDirectory は作業ディレクトリの入力を求める。
func promptWorkingDirectory() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	defaultDir := "C:/fixup-commit-sync-manager"

	fmt.Printf("作業ディレクトリを指定してください [%s]: ", defaultDir)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	workDir := strings.TrimSpace(input)
	if workDir == "" {
		workDir = defaultDir
	}

	// パス区切り文字を正規化。
	workDir = filepath.FromSlash(workDir)

	return workDir, nil
}

// createWorkingDirectory は作業ディレクトリを作成する。
func createWorkingDirectory(workDir string) error {
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		fmt.Printf("    作業ディレクトリを作成しています: %s\n", workDir)
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return err
		}
		fmt.Println("    作業ディレクトリを作成しました。")
	} else {
		fmt.Printf("    作業ディレクトリは既に存在します: %s\n", workDir)
	}
	return nil
}

// copyExecutable は実行ファイルを作業ディレクトリにコピーする。
func copyExecutable(workDir string) error {
	// 現在の実行ファイルのパスを取得。
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execName := filepath.Base(execPath)
	targetPath := filepath.Join(workDir, execName)

	// 既に存在する場合はスキップ。
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("    実行ファイルは既に存在します: %s\n", targetPath)
		return nil
	}

	fmt.Printf("    実行ファイルをコピーしています: %s -> %s\n", execPath, targetPath)

	// ファイルをコピー。
	input, err := os.ReadFile(execPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(targetPath, input, 0755); err != nil {
		return err
	}

	fmt.Println("    実行ファイルをコピーしました。")
	return nil
}

// createVHDXFile はVHDXファイルを作成する。
func createVHDXFile(vhdxPath string) error {
	// 既に存在する場合はスキップ。
	if _, err := os.Stat(vhdxPath); err == nil {
		fmt.Printf("VHDXファイルは既に存在します: %s\n", vhdxPath)
		return nil
	}

	fmt.Printf("VHDXファイルを作成しています: %s\n", vhdxPath)

	// VHDXマネージャーを使用して実際のVHDXファイルを作成。
	vhdxManager := vhdx.NewManager(vhdxPath, "X:")
	
	// VHDXファイルを作成（10GBデフォルト、暗号化なし）。
	err := vhdxManager.Create("10GB", false)
	if err != nil {
		// 実際のVHDX作成が失敗した場合は、ダミーファイルを作成。
		fmt.Printf("実際のVHDX作成に失敗しました。ダミーファイルを作成します: %v\n", err)
		
		// ディレクトリを作成。
		if err := os.MkdirAll(filepath.Dir(vhdxPath), 0755); err != nil {
			return fmt.Errorf("VHDXディレクトリ作成エラー: %v", err)
		}
		
		// ダミーファイルを作成。
		file, err := os.Create(vhdxPath)
		if err != nil {
			return fmt.Errorf("ダミーVHDXファイル作成エラー: %v", err)
		}
		file.Close()
		
		fmt.Println("VHDXダミーファイルの準備が完了しました。")
		fmt.Println("注意: 実際のVHDX初期化は init-vhdx コマンドで行ってください。")
		return nil
	}

	fmt.Println("VHDXファイルの作成が完了しました。")
	return nil
}

// generateInitialConfig は初期設定ファイルを生成する。
func generateInitialConfig(workDir, configPath string) error {
	fmt.Println("\n    === 設定ファイルの生成 ===")
	fmt.Println("    対話的に設定ファイルを作成します。")
	fmt.Println()

	// 改善されたinit-configを呼び出す。
	return runInitConfigImproved(configPath, workDir)
}