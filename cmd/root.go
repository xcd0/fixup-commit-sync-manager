package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "FixupCommitSyncManager",
	Short: "Dev リポジトリと Ops リポジトリ間でソースファイルを同期するツール",
	Long: `FixupCommitSyncManager は Windows 環境向けの総合運用プラットフォームです:

- Dev リポジトリ⇔Ops リポジトリ間のソースファイル自動同期
- VHDX を用いた隔離初期化機能
- 設定ファイルの生成と検証
- autosquash 対応の自動 fixup コミット機能
- 包括的なログ記録とエラーハンドリング

利用可能なサブコマンド:
- init-config      : 対話型ウィザードで設定ファイルを作成
- validate-config  : 設定ファイルの構文と内容を検証
- init-vhdx        : VHDX ファイルを初期化して Ops リポジトリをセットアップ
- mount-vhdx       : VHDX ファイルをマウント
- unmount-vhdx     : VHDX ファイルをアンマウント
- snapshot-vhdx    : VHDX スナップショットを管理
- sync             : リポジトリ間でファイルを同期
- fixup            : fixup コミットを実行`,
	Version: "1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "設定ファイルのパス (デフォルト: config.hjson)")
	rootCmd.PersistentFlags().Bool("dry-run", false, "実際の変更を行わずにプレビュー実行")
	rootCmd.PersistentFlags().Bool("verbose", false, "詳細な出力を有効化")

	rootCmd.AddCommand(NewInitConfigCmd())
	rootCmd.AddCommand(NewValidateConfigCmd())
	rootCmd.AddCommand(NewInitVHDXCmd())
	rootCmd.AddCommand(NewMountVHDXCmd())
	rootCmd.AddCommand(NewUnmountVHDXCmd())
	rootCmd.AddCommand(NewSnapshotVHDXCmd())
	rootCmd.AddCommand(NewSyncCmd())
	rootCmd.AddCommand(NewFixupCmd())
	rootCmd.AddCommand(NewHelpCmd())
}

func NewHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "サブコマンドのヘルプ情報を表示",
		Long:  "利用可能なすべてのサブコマンドの詳細ヘルプ情報を表示します",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				rootCmd.Help()
				return
			}

			subCmd, _, err := rootCmd.Find(args)
			if err != nil {
				fmt.Printf("不明なコマンド: %s\n", args[0])
				rootCmd.Help()
				return
			}

			subCmd.Help()
		},
	}
}
