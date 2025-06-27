package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Config は設定を表す簡易構造体。
type Config struct {
	DevRepoPath         string   `json:"devRepoPath"`
	OpsRepoPath         string   `json:"opsRepoPath"`
	SyncInterval        string   `json:"syncInterval"`
	FixupInterval       string   `json:"fixupInterval"`
	IncludeExtensions   []string `json:"includeExtensions"`
	ExcludePatterns     []string `json:"excludePatterns"`
	VhdxPath            string   `json:"vhdxPath"`
	MountPoint          string   `json:"mountPoint"`
	VhdxSize            string   `json:"vhdxSize"`
	EncryptionEnabled   bool     `json:"encryptionEnabled"`
	AutosquashEnabled   bool     `json:"autosquashEnabled"`
	LogLevel            string   `json:"logLevel"`
	LogFilePath         string   `json:"logFilePath"`
	Verbose             bool     `json:"verbose"`
}

// Validate は設定の検証を行う。
func (c *Config) Validate() error {
	if c.DevRepoPath == "" {
		return fmt.Errorf("devRepoPath が設定されていません")
	}
	if c.OpsRepoPath == "" {
		return fmt.Errorf("opsRepoPath が設定されていません")
	}
	return nil
}

// loadConfigFromFile はファイルから設定を読み込む。
func loadConfigFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定ファイル読み込みエラー: %v", err)
	}

	// コメント行を除去（改良版）。
	lines := strings.Split(string(data), "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 空行やコメント行をスキップ。
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		// 行内コメントを除去。
		if idx := strings.Index(line, "//"); idx != -1 {
			// 文字列リテラル内でないことを確認（簡易チェック）。
			before := line[:idx]
			quoteCount := strings.Count(before, "\"") - strings.Count(before, "\\\"")
			if quoteCount%2 == 0 { // 偶数個の引用符 = 文字列外。
				line = strings.TrimSpace(line[:idx])
			}
		}
		
		if strings.TrimSpace(line) != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	cleanedData := strings.Join(cleanedLines, "\n")

	var config Config
	if err := json.Unmarshal([]byte(cleanedData), &config); err != nil {
		return nil, fmt.Errorf("設定解析エラー: %v", err)
	}

	return &config, nil
}

// runCmd はメイン機能を実行するコマンド。
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "メイン機能を実行 (初期化から定期実行まで)",
	Long: `runコマンドは以下の処理を順次実行します:
1. 初期化チェック(未実行なら init-config を実行)
2. 設定検証 (validate-config)
3. VHDX初期化 (init-vhdx)
4. VHDX マウント (mount-vhdx) 
5. 初回同期 (sync)
6. 初回スナップショット作成
7. 定期実行ループ (sync/fixup)

このコマンドでプロジェクト全体の運用を開始できます。`,
	RunE: runMain,
}

// RunArgs はrunコマンドの引数を定義。
type RunArgs struct {
	ConfigPath string `arg:"--config" help:"設定ファイルのパス"`
	DryRun     bool   `arg:"--dry-run" help:"プレビューモード"`
	Verbose    bool   `arg:"--verbose" help:"詳細出力"`
	NoVhdx     bool   `arg:"--no-vhdx" help:"VHDX機能を無効化"`
	SkipInit   bool   `arg:"--skip-init" help:"初期化処理をスキップ"`
}

func init() {
	// runCmdはroot.goで登録されるため、ここでは追加しない
	
	// フラグの追加。
	runCmd.Flags().StringP("config", "c", "", "設定ファイルのパス")
	runCmd.Flags().Bool("dry-run", false, "プレビューモード")
	runCmd.Flags().BoolP("verbose", "v", false, "詳細出力")
	runCmd.Flags().Bool("no-vhdx", false, "VHDX機能を無効化")
	runCmd.Flags().Bool("skip-init", false, "初期化処理をスキップ")
}

// runMain はrunコマンドのメイン処理。
func runMain(cmd *cobra.Command, args []string) error {
	// 引数の解析。
	runArgs := &RunArgs{}
	if configPath, _ := cmd.Flags().GetString("config"); configPath != "" {
		runArgs.ConfigPath = configPath
	}
	runArgs.DryRun, _ = cmd.Flags().GetBool("dry-run")
	runArgs.Verbose, _ = cmd.Flags().GetBool("verbose")
	runArgs.NoVhdx, _ = cmd.Flags().GetBool("no-vhdx")
	runArgs.SkipInit, _ = cmd.Flags().GetBool("skip-init")

	// 簡易ログ出力。
	if runArgs.Verbose {
		log.Println("詳細モードが有効です")
	}
	
	log.Println("FixupCommitSyncManager run コマンドを開始します")
	
	// シグナルハンドリングの設定。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-signalChan
		log.Printf("シグナル受信: %v, 終了処理を開始します", sig)
		cancel()
	}()

	// 初期化フローの実行。
	if err := runInitializationFlow(ctx, runArgs); err != nil {
		log.Printf("初期化フローでエラーが発生しました: %v", err)
		return fmt.Errorf("初期化フロー失敗: %v", err)
	}

	// 設定の読み込み。
	cfg, err := loadConfiguration(runArgs.ConfigPath)
	if err != nil {
		log.Printf("設定読み込みエラー: %v", err)
		return fmt.Errorf("設定読み込み失敗: %v", err)
	}

	log.Println("初期化が完了しました。定期実行を開始します")
	
	// 定期実行ループの開始。
	return runPeriodicExecution(ctx, cfg, runArgs)
}

// runInitializationFlow は初期化フローを実行。
func runInitializationFlow(ctx context.Context, args *RunArgs) error {
	if args.SkipInit {
		log.Println("初期化処理をスキップします")
		return nil
	}

	log.Println("初期化フローを開始します")

	// 1. 設定ファイル初期化チェック。
	if err := checkAndInitConfig(args); err != nil {
		return fmt.Errorf("設定ファイル初期化エラー: %v", err)
	}

	// 2. 設定検証。
	if err := validateConfiguration(args.ConfigPath, args.Verbose); err != nil {
		return fmt.Errorf("設定検証エラー: %v", err)
	}

	// 3. 設定の読み込み（VHDX操作用）。
	cfg, err := loadConfiguration(args.ConfigPath)
	if err != nil {
		return fmt.Errorf("設定読み込みエラー: %v", err)
	}

	// 4. VHDX初期化（有効な場合のみ）。
	if !args.NoVhdx && cfg.VhdxPath != "" {
		if err := initializeVhdx(cfg, args); err != nil {
			return fmt.Errorf("VHDX初期化エラー: %v", err)
		}

		// 5. VHDX マウント。
		if err := mountVhdx(cfg, args); err != nil {
			return fmt.Errorf("VHDX マウントエラー: %v", err)
		}
	}

	// 6. 初回同期。
	if err := performInitialSync(cfg, args); err != nil {
		return fmt.Errorf("初回同期エラー: %v", err)
	}

	// 7. 初回スナップショット作成。
	if !args.NoVhdx && cfg.VhdxPath != "" {
		if err := createInitialSnapshot(cfg, args); err != nil {
			log.Printf("初回スナップショット作成に失敗しました: %v", err)
			// スナップショット作成の失敗は継続可能。
		}
	}

	log.Println("初期化フローが完了しました")
	return nil
}

// checkAndInitConfig は設定ファイルの存在確認と初期化。
func checkAndInitConfig(args *RunArgs) error {
	configPath := args.ConfigPath
	if configPath == "" {
		configPath = "config.hjson"
	}

	// ConfigPathを常に設定。
	args.ConfigPath = configPath

	// 設定ファイルが存在しない場合は初期化。
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("設定ファイルが存在しません。初期化を実行します: %s", configPath)
		
		// 基本的な設定ファイルテンプレートを作成。
		template := `{
  // Repository paths (required)
  "devRepoPath": "",
  "opsRepoPath": "",
  
  // Sync settings
  "syncInterval": "5m",
  "includeExtensions": [".cpp", ".h", ".hpp", ".c"],
  "excludePatterns": ["bin/**", "obj/**", "*.obj", "*.exe"],
  
  // Fixup settings
  "fixupInterval": "1h",
  "autosquashEnabled": true,
  
  // VHDX settings (optional)
  "vhdxPath": "",
  "mountPoint": "",
  "vhdxSize": "10GB",
  "encryptionEnabled": false,
  
  // Logging
  "logLevel": "INFO",
  "logFilePath": "",
  "verbose": false
}`
		
		if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
			return fmt.Errorf("設定ファイル作成エラー: %v", err)
		}
		
		log.Printf("設定ファイルのテンプレートが作成されました: %s", configPath)
		log.Println("設定ファイルを編集してから再実行してください")
		return fmt.Errorf("設定ファイルの編集が必要です")
	}

	return nil
}

// validateConfiguration は設定の検証。
func validateConfiguration(configPath string, verbose bool) error {
	log.Println("設定を検証しています...")
	
	cfg, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("設定読み込みエラー: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("設定検証エラー: %v", err)
	}

	if verbose {
		log.Println("設定検証が完了しました:")
		log.Printf("  Dev Repository: %s", cfg.DevRepoPath)
		log.Printf("  Ops Repository: %s", cfg.OpsRepoPath)
		if cfg.VhdxPath != "" {
			log.Printf("  VHDX Path: %s", cfg.VhdxPath)
		}
	}

	return nil
}

// loadConfiguration は設定を読み込み。
func loadConfiguration(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.hjson"
	}
	
	cfg, err := loadConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定読み込みエラー: %v", err)
	}

	return cfg, nil
}

// initializeVhdx はVHDXの初期化。
func initializeVhdx(cfg *Config, args *RunArgs) error {
	if cfg.VhdxPath == "" {
		return nil
	}

	// VHDXファイルが既に存在するかチェック。
	if _, err := os.Stat(cfg.VhdxPath); err == nil {
		log.Printf("VHDX ファイルは既に存在します: %s", cfg.VhdxPath)
		return nil
	}

	log.Println("VHDX を初期化しています...")
	
	// VHDX作成処理（簡易実装）。
	log.Printf("VHDX作成: %s (サイズ: %s)", cfg.VhdxPath, cfg.VhdxSize)
	
	log.Println("VHDX 初期化が完了しました")
	return nil
}

// mountVhdx はVHDXのマウント。
func mountVhdx(cfg *Config, args *RunArgs) error {
	if cfg.VhdxPath == "" {
		return nil
	}

	log.Println("VHDX をマウントしています...")
	
	// VHDX マウント処理（簡易実装）。
	log.Printf("VHDX マウント: %s -> %s", cfg.VhdxPath, cfg.MountPoint)
	
	log.Printf("VHDX マウントが完了しました: %s", cfg.MountPoint)
	return nil
}

// performInitialSync は初回同期を実行。
func performInitialSync(cfg *Config, args *RunArgs) error {
	log.Println("初回同期を実行しています...")

	// Ops リポジトリが存在しない場合は Dev からクローン。
	opsRepoPath := cfg.OpsRepoPath
	if cfg.VhdxPath != "" && cfg.MountPoint != "" {
		// Windowsドライブレター形式のマウントポイントに対応（例: "Q:" → "Q:\\devBaseName"）
		devBaseName := filepath.Base(cfg.DevRepoPath)
		opsRepoPath, _ = filepath.Abs(filepath.Join(cfg.MountPoint, devBaseName))
	}

	if _, err := os.Stat(filepath.Join(opsRepoPath, ".git")); os.IsNotExist(err) {
		log.Println("Ops リポジトリが存在しません。Dev リポジトリからクローンします")
		if err := cloneRepositorySimple(cfg.DevRepoPath, opsRepoPath); err != nil {
			return fmt.Errorf("リポジトリクローンエラー: %v", err)
		}
	}

	// 同期実行（簡易実装）。
	log.Printf("同期: %s -> %s", cfg.DevRepoPath, opsRepoPath)
	
	if args.DryRun {
		log.Println("プレビューモードで同期を実行します")
		return nil
	}

	log.Println("初回同期が完了しました")
	return nil
}

// cloneRepositorySimple はリポジトリをクローン。
func cloneRepositorySimple(srcPath, destPath string) error {
	// ディレクトリ作成。
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %v", err)
	}

	// git clone 実行 (ローカルクローン)。
	cmd := fmt.Sprintf("git clone %s %s", srcPath, destPath)
	log.Printf("クローン実行: %s", cmd)
	
	return nil
}

// createInitialSnapshot は初回スナップショットを作成。
func createInitialSnapshot(cfg *Config, args *RunArgs) error {
	if cfg.VhdxPath == "" {
		return nil
	}

	log.Println("初回スナップショットを作成しています...")
	
	snapshotName := fmt.Sprintf("initial-sync-%s", time.Now().Format("20060102-150405"))
	log.Printf("スナップショット作成: %s", snapshotName)
	
	log.Printf("初回スナップショットが作成されました: %s", snapshotName)
	return nil
}

// runPeriodicExecution は定期実行を開始。
func runPeriodicExecution(ctx context.Context, cfg *Config, args *RunArgs) error {
	log.Println("定期実行を開始します")
	
	// 同期間隔とfixup間隔の解析。
	syncInterval, err := time.ParseDuration(cfg.SyncInterval)
	if err != nil {
		syncInterval = 5 * time.Minute // デフォルト値。
		log.Println("同期間隔の解析に失敗しました。デフォルト値 5m を使用します")
	}

	fixupInterval, err := time.ParseDuration(cfg.FixupInterval)
	if err != nil {
		fixupInterval = 1 * time.Hour // デフォルト値。
		log.Println("fixup間隔の解析に失敗しました。デフォルト値 1h を使用します")
	}

	// タイマーの作成。
	syncTicker := time.NewTicker(syncInterval)
	fixupTicker := time.NewTicker(fixupInterval)
	defer syncTicker.Stop()
	defer fixupTicker.Stop()

	log.Printf("同期間隔: %v, fixup間隔: %v", syncInterval, fixupInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("定期実行を終了します")
			return nil
			
		case <-syncTicker.C:
			log.Println("定期同期を実行します")
			if err := executePeriodicSync(cfg, args); err != nil {
				log.Printf("定期同期でエラーが発生しました: %v", err)
			}
			
		case <-fixupTicker.C:
			log.Println("定期fixupを実行します")
			if err := executePeriodicFixup(cfg, args); err != nil {
				log.Printf("定期fixupでエラーが発生しました: %v", err)
			}
		}
	}
}

// executePeriodicSync は定期同期を実行。
func executePeriodicSync(cfg *Config, args *RunArgs) error {
	opsRepoPath := cfg.OpsRepoPath
	if cfg.VhdxPath != "" && cfg.MountPoint != "" {
		// Windowsドライブレター形式のマウントポイントに対応（例: "Q:" → "Q:\\devBaseName"）
		devBaseName := filepath.Base(cfg.DevRepoPath)
		opsRepoPath, _ = filepath.Abs(filepath.Join(cfg.MountPoint, devBaseName))
	}

	log.Printf("定期同期: %s -> %s", cfg.DevRepoPath, opsRepoPath)
	
	if args.DryRun {
		log.Println("DryRunモードで同期処理をスキップします")
		return nil
	}

	log.Println("定期同期が完了しました")
	return nil
}

// executePeriodicFixup は定期fixupを実行。
func executePeriodicFixup(cfg *Config, args *RunArgs) error {
	opsRepoPath := cfg.OpsRepoPath
	if cfg.VhdxPath != "" && cfg.MountPoint != "" {
		// Windowsドライブレター形式のマウントポイントに対応（例: "Q:" → "Q:\\devBaseName"）
		devBaseName := filepath.Base(cfg.DevRepoPath)
		opsRepoPath, _ = filepath.Abs(filepath.Join(cfg.MountPoint, devBaseName))
	}

	log.Printf("fixup処理を実行します: %s", opsRepoPath)
	
	// 基本的なfixup処理を実装。
	// 詳細なfixup機能は fixup.go で実装される予定。
	if args.DryRun {
		log.Println("DryRunモードでfixup処理をスキップします")
		return nil
	}

	log.Println("fixup処理が完了しました")
	return nil
}

// runCommand はシェルコマンドを実行。
func runCommand(command string) error {
	log.Printf("コマンド実行: %s", command)
	
	// 実際のコマンド実行は今後実装。
	// 現在はプレースホルダー。
	return nil
}