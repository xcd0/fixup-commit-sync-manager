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
- run              : メイン機能を実行（初期化から定期実行まで一括処理）
- init             : 初期セットアップ（作業ディレクトリ作成、設定生成）
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

	// ヘルプテンプレートを日本語化
	rootCmd.SetUsageTemplate(getUsageTemplate())
	rootCmd.SetHelpTemplate(getHelpTemplate())

	// completion コマンドを無効化
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	
	// デフォルトのhelpコマンドを無効化して日本語版を追加
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [コマンド]",
		Short:  "コマンドのヘルプを表示",
		Long:   "指定されたコマンドの詳細なヘルプ情報を表示します。",
		Hidden: true,
	})

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(NewInitConfigCmd())
	rootCmd.AddCommand(NewValidateConfigCmd())
	rootCmd.AddCommand(NewInitVHDXCmd())
	rootCmd.AddCommand(NewMountVHDXCmd())
	rootCmd.AddCommand(NewUnmountVHDXCmd())
	rootCmd.AddCommand(NewSnapshotVHDXCmd())
	rootCmd.AddCommand(NewSyncCmd())
	rootCmd.AddCommand(NewFixupCmd())
}

// getUsageTemplate は日本語化されたUsageテンプレートを返す。
func getUsageTemplate() string {
	return `使用法:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [コマンド]{{end}}{{if gt (len .Aliases) 0}}

別名:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

例:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

利用可能なコマンド:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

オプション:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

グローバルオプション:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

追加ヘルプトピック:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

詳細は "{{.CommandPath}} [コマンド] --help" を実行してください。{{end}}
`
}

// getHelpTemplate は日本語化されたHelpテンプレートを返す。
func getHelpTemplate() string {
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
}
