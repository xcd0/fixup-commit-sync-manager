package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// NewCompletionCmd はカスタム補完コマンドを作成する。
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "シェル補完スクリプトを生成してインストール",
		Long: `シェル補完スクリプトを生成し、適切なユーザーディレクトリに配置します。

引数を省略した場合、実行中のシェルを自動判別します。

サポートされているシェル: bash, zsh, fish, powershell

インストール先:
- bash: $HOME/.bash_completion
- zsh:  $HOME/.zsh/completion/_fixup-commit-sync-manager
- fish: $HOME/.config/fish/completions/fixup-commit-sync-manager.fish
- powershell: $HOME/Documents/PowerShell/Scripts/fixup-completion.ps1`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE:      runCompletion,
	}

	cmd.Flags().Bool("install", true, "補完スクリプトをユーザーディレクトリに自動インストール")
	cmd.Flags().Bool("print", false, "スクリプトを標準出力に表示（インストールしない）")

	return cmd
}

// runCompletion は補完スクリプトの生成とインストールを実行する。
func runCompletion(cmd *cobra.Command, args []string) error {
	var shell string
	
	if len(args) > 0 {
		shell = args[0]
	} else {
		// シェル自動判別
		var err error
		shell, err = detectShell()
		if err != nil {
			return fmt.Errorf("シェルの自動判別に失敗しました: %w", err)
		}
		fmt.Printf("検出されたシェル: %s\n", shell)
	}

	printOnly, _ := cmd.Flags().GetBool("print")
	install, _ := cmd.Flags().GetBool("install")

	if printOnly {
		return generateCompletionScript(shell, "")
	}

	if install {
		installPath, err := getInstallPath(shell)
		if err != nil {
			return fmt.Errorf("インストールパスの取得に失敗しました: %w", err)
		}

		fmt.Printf("補完スクリプトをインストール中: %s\n", installPath)
		
		// ディレクトリを作成
		if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
		}

		if err := generateCompletionScript(shell, installPath); err != nil {
			return err
		}

		fmt.Printf("✓ 補完スクリプトがインストールされました: %s\n", installPath)
		printActivationInstructions(shell, installPath)
		return nil
	}

	// デフォルトは標準出力
	return generateCompletionScript(shell, "")
}

// detectShell は実行中のシェルを自動判別する。
func detectShell() (string, error) {
	// SHELL環境変数から判別
	if shell := os.Getenv("SHELL"); shell != "" {
		baseName := filepath.Base(shell)
		switch baseName {
		case "bash":
			return "bash", nil
		case "zsh":
			return "zsh", nil
		case "fish":
			return "fish", nil
		}
	}

	// PowerShellの判別
	if runtime.GOOS == "windows" {
		if psModulePath := os.Getenv("PSModulePath"); psModulePath != "" {
			return "powershell", nil
		}
	}

	// 親プロセス名から判別
	if ppid := os.Getppid(); ppid > 1 {
		// Linux/macOSでの簡易判別
		if runtime.GOOS != "windows" {
			if procName, err := getProcessName(ppid); err == nil {
				switch {
				case strings.Contains(procName, "bash"):
					return "bash", nil
				case strings.Contains(procName, "zsh"):
					return "zsh", nil
				case strings.Contains(procName, "fish"):
					return "fish", nil
				}
			}
		}
	}

	// デフォルトはbash
	return "bash", nil
}

// getProcessName は指定されたPIDのプロセス名を取得する。
func getProcessName(pid int) (string, error) {
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("Windows での詳細なプロセス名取得は未対応")
	}

	// Linux/macOS での /proc/PID/comm を読み取り
	procPath := fmt.Sprintf("/proc/%d/comm", pid)
	if data, err := os.ReadFile(procPath); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("プロセス名の取得に失敗")
}

// getInstallPath は各シェルの補完スクリプトインストールパスを返す。
func getInstallPath(shell string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	switch shell {
	case "bash":
		return filepath.Join(homeDir, ".bash_completion"), nil
	case "zsh":
		return filepath.Join(homeDir, ".zsh", "completion", "_fixup-commit-sync-manager"), nil
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "completions", "fixup-commit-sync-manager.fish"), nil
	case "powershell":
		if runtime.GOOS == "windows" {
			// Windows PowerShell 5.x の場合
			if psPath := os.Getenv("USERPROFILE"); psPath != "" {
				return filepath.Join(psPath, "Documents", "WindowsPowerShell", "Scripts", "fixup-completion.ps1"), nil
			}
			// PowerShell Core (pwsh) の場合
			return filepath.Join(homeDir, "Documents", "PowerShell", "Scripts", "fixup-completion.ps1"), nil
		} else {
			// Linux/macOS での PowerShell Core
			return filepath.Join(homeDir, ".config", "powershell", "Scripts", "fixup-completion.ps1"), nil
		}
	default:
		return "", fmt.Errorf("サポートされていないシェル: %s", shell)
	}
}

// generateCompletionScript は指定されたシェル用の補完スクリプトを生成する。
func generateCompletionScript(shell, outputPath string) error {
	var err error

	if outputPath == "" {
		// 標準出力に出力
		switch shell {
		case "bash":
			err = rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			err = rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			err = rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("サポートされていないシェル: %s", shell)
		}
	} else {
		// ファイルに出力
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("ファイルの作成に失敗: %w", err)
		}
		defer file.Close()

		switch shell {
		case "bash":
			err = rootCmd.GenBashCompletion(file)
		case "zsh":
			err = rootCmd.GenZshCompletion(file)
		case "fish":
			err = rootCmd.GenFishCompletion(file, true)
		case "powershell":
			err = rootCmd.GenPowerShellCompletionWithDesc(file)
		default:
			return fmt.Errorf("サポートされていないシェル: %s", shell)
		}
	}

	return err
}

// printActivationInstructions は補完スクリプトの有効化手順を表示する。
func printActivationInstructions(shell, installPath string) {
	fmt.Println("\n=== 補完スクリプトの有効化手順 ===")
	
	switch shell {
	case "bash":
		fmt.Printf("以下のコマンドを実行して補完を有効化してください:\n")
		fmt.Printf("  echo 'source %s' >> ~/.bashrc\n", installPath)
		fmt.Printf("  source ~/.bashrc\n")
	case "zsh":
		fmt.Printf("以下のディレクトリがfpathに含まれていることを確認してください:\n")
		fmt.Printf("  echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc\n")
		fmt.Printf("  echo 'autoload -U compinit && compinit' >> ~/.zshrc\n")
		fmt.Printf("  source ~/.zshrc\n")
	case "fish":
		fmt.Printf("Fishの補完は自動的に有効になります。新しいシェルセッションを開始してください。\n")
	case "powershell":
		fmt.Printf("PowerShellプロファイルに以下を追加してください:\n")
		if runtime.GOOS == "windows" {
			fmt.Printf("  # PowerShell起動時に以下のコマンドを実行:\n")
			fmt.Printf("  if (!(Test-Path -Path $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force }\n")
			fmt.Printf("  Add-Content -Path $PROFILE -Value \". '%s'\"\n", strings.ReplaceAll(installPath, "\\", "/"))
			fmt.Printf("  \n")
			fmt.Printf("  # または手動でプロファイルを編集:\n")
			fmt.Printf("  notepad $PROFILE\n")
		} else {
			profilePath := "~/.config/powershell/Microsoft.PowerShell_profile.ps1"
			fmt.Printf("  echo '. %s' >> %s\n", installPath, profilePath)
		}
	}
	
	fmt.Println("\n補完が有効になった後、タブキーで補完候補が表示されます。")
}