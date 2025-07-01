# FixupCommitSyncManager

Windows 環境向けの Dev リポジトリ⇔Ops リポジトリ間ソースファイル自動同期ツール

## 概要

FixupCommitSyncManager は、Windows 環境で開発リポジトリと運用リポジトリ間のソースファイルを軽量に自動同期し、定期的に fixup コミットを行う総合運用プラットフォームです。

### 主な機能

- **🔄 動的ブランチ追従**: Dev リポジトリのカレントブランチに自動追従
- **📁 自動ファイル同期**: Dev リポジトリ⇔Ops リポジトリ間のソースファイル自動同期
- **🌟 ブランチ自動管理**: 必要に応じてブランチを自動作成・切り替え
- **💾 VHDX サポート**: VHDX を用いた隔離初期化機能
- **⚙️ 設定管理**: 対話型ウィザードによる設定ファイル生成と検証（ブランチ設定不要）
- **🔄 Fixup コミット**: autosquash 対応の自動 fixup コミット機能
- **📊 包括的ログ**: 構造化ログとエラー通知システム

## インストール

### バイナリダウンロード

[Releases](https://github.com/your-org/FixupCommitSyncManager/releases) から最新版をダウンロード

### ソースからビルド

```bash
git clone https://github.com/your-org/FixupCommitSyncManager.git
cd FixupCommitSyncManager
make build
```

## クイックスタート

### 1. 設定ファイル作成

```bash
# 対話型ウィザードで設定ファイルを作成
./fixup-commit-sync-manager init-config

# 設定ファイルを検証
./fixup-commit-sync-manager validate-config --verbose
```

### 2. VHDX 初期化（オプション）

```bash
# VHDX ファイルを作成し、Ops リポジトリをセットアップ
./fixup-commit-sync-manager init-vhdx
```

### 3. ファイル同期実行

```bash
# 一回のみ同期（Dev のカレントブランチに自動追従）
./fixup-commit-sync-manager sync

# 継続的同期（5分間隔、ブランチ変更も自動検出）
./fixup-commit-sync-manager sync --continuous
```

### 4. Fixup コミット実行

```bash
# 一回のみ fixup（Dev のカレントブランチで実行）
./fixup-commit-sync-manager fixup

# 継続的 fixup（1時間間隔、ブランチ変更も自動追従）
./fixup-commit-sync-manager fixup --continuous
```

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `init-config` | 対話型ウィザードで設定ファイルを作成（ブランチ設定不要） |
| `validate-config` | 設定ファイルの構文と内容を検証 |
| `sync` | Dev↔Ops リポジトリ間でファイルを動的ブランチ追従で同期 |
| `fixup` | 動的ブランチ追従で fixup コミットを実行 |
| `init-vhdx` | VHDX ファイルを初期化 |
| `mount-vhdx` | VHDX ファイルをマウント |
| `unmount-vhdx` | VHDX ファイルをアンマウント |
| `snapshot-vhdx` | VHDX スナップショットを管理 |
| `completion` | シェル補完スクリプトを生成 |

### シェル補完

Bash、Zsh、Fish、PowerShell でのタブ補完をサポートしています。

```bash
# 実行中のシェルを自動判別して補完スクリプトをインストール
./fixup-commit-sync-manager completion

# 特定のシェル用にインストール
./fixup-commit-sync-manager completion bash
./fixup-commit-sync-manager completion zsh
./fixup-commit-sync-manager completion fish
./fixup-commit-sync-manager completion powershell

# スクリプトを標準出力に表示（インストールせずに確認）
./fixup-commit-sync-manager completion --print
```

**自動インストール先:**
- **Bash**: `$HOME/.bash_completion`
- **Zsh**: `$HOME/.zsh/completion/_fixup-commit-sync-manager`
- **Fish**: `$HOME/.config/fish/completions/fixup-commit-sync-manager.fish`
- **PowerShell**: `$HOME/Documents/PowerShell/Scripts/fixup-completion.ps1`

### グローバルオプション

- `--config <path>`: 設定ファイルのパス（デフォルト: config.hjson）
- `--dry-run`: 実際の変更を行わずにプレビュー実行
- `--verbose`: 詳細な出力を有効化

## 設定ファイル例

```hjson
{
  // === Repository Settings ===
  "devRepoPath": "C:\\path\\to\\dev-repo",
  "opsRepoPath": "C:\\path\\to\\ops-repo",

  // === Sync Settings ===
  "syncInterval": "5m",
  "includeExtensions": [".cpp", ".h", ".hpp"],
  "excludePatterns": ["bin/**", "obj/**"],

  // === Fixup Settings ===
  "fixupInterval": "1h",
  "autosquashEnabled": true,
  // Note: Branch settings are now dynamic - automatically tracks Dev repository's current branch

  // === VHDX Settings ===
  "vhdxPath": "C:\\vhdx\\ops.vhdx",
  "mountPoint": "X:",
  "vhdxSize": "10GB",

  // === Logging ===
  "logLevel": "INFO",
  "logFilePath": "C:\\logs\\sync.log"
}
```

## 使用例

### 基本的なワークフロー

```bash
# 1. 設定ファイル作成（ブランチ設定は不要）
./fixup-commit-sync-manager init-config

# 2. 設定検証
./fixup-commit-sync-manager validate-config --verbose

# 3. 一回のみ同期（Dev のカレントブランチに自動追従）
./fixup-commit-sync-manager sync --verbose

# 4. 継続的運用開始（ブランチ変更も自動検出）
./fixup-commit-sync-manager sync --continuous &
./fixup-commit-sync-manager fixup --continuous &
```

### 動的ブランチ追従の例

```bash
# Dev 側で feature-abc ブランチに切り替え
cd /path/to/dev-repo
git checkout feature-abc

# 同期実行 → Ops 側も自動的に feature-abc ブランチに切り替わる
./fixup-commit-sync-manager sync

# Dev 側で main ブランチに戻る
cd /path/to/dev-repo  
git checkout main

# 同期実行 → Ops 側も自動的に main ブランチに切り替わる
./fixup-commit-sync-manager sync
```

### VHDX を使用したワークフロー

```bash
# 1. VHDX 初期化
./fixup-commit-sync-manager init-vhdx

# 2. スナップショット作成
./fixup-commit-sync-manager snapshot-vhdx create before-sync

# 3. 同期実行
./fixup-commit-sync-manager sync

# 4. 問題があればロールバック
./fixup-commit-sync-manager snapshot-vhdx rollback before-sync
```

## 開発

### 前提条件

- Go 1.20+
- Git
- Windows 10/11（VHDX 機能使用時）

### ビルド

```bash
# デバッグビルド
make build

# リリースビルド
make release

# テスト実行
make test

# リント実行
make lint
```

### 開発コマンド

```bash
# プロジェクト情報表示
make info

# デモ実行
make demo

# ファイル変更監視
make watch

# テストカバレッジ
make test-coverage
```

## テスト

FixupCommitSyncManager は包括的なテストスイートを提供しており、様々なレベルでの品質保証を行っています。

### テストカテゴリ

#### 📝 **基本テスト**
- **ユニットテスト**: 各パッケージの個別機能テスト
- **統合テスト**: コンポーネント間の連携テスト
- **E2Eテスト**: 実際のGitリポジトリを使った統合テスト

#### 🔄 **E2E統合テスト**
- **コマンド実行テスト**: 基本的なコマンド実行機能
- **実際のワークフローテスト**: 本物のGitリポジトリでの動作確認
- **完全ワークフローテスト**: 30秒間の継続的開発シミュレーション

### テストコマンド一覧

#### **基本テスト**
```bash
make test              # 全テスト実行
make test-short        # 短時間テスト（統合テスト除く）
make test-unit         # ユニットテストのみ
```

#### **統合テスト**
```bash
make test-integration  # 既存統合テスト
make test-e2e          # E2E統合テスト
make test-full         # 全統合テスト
```

#### **パッケージ別テスト**
```bash
make test-cmd          # CMDパッケージ
make test-config       # Configパッケージ
make test-sync         # Syncパッケージ
make test-fixup        # Fixupパッケージ
make test-vhdx         # VHDXパッケージ
```

#### **E2E個別テスト**
```bash
make test-command-execution # コマンド実行テスト
make test-real-workflow     # 実際のワークフロー
make test-complete-workflow # 完全ワークフロー（30秒）
```

#### **品質保証フロー**
```bash
make test-coverage         # テストカバレッジ
make test-coverage-detail  # 詳細カバレッジ（パッケージ別）
make test-ci              # CI用フロー（短時間）
make test-qa              # QA用フロー（包括的）
make test-release         # リリース前フロー
```

#### **便利コマンド**
```bash
make test-summary     # 全テストコマンドの概要表示
```

### 開発フローでの使用例

#### **開発中の素早いチェック**
```bash
# 短時間での基本チェック
make test-short

# 特定パッケージの詳細チェック
make test-sync
make test-config
```

#### **機能追加時の確認**
```bash
# 関連パッケージのテスト
make test-sync test-fixup

# E2E統合テストで動作確認
make test-command-execution
make test-real-workflow
```

#### **リリース前の品質確認**
```bash
# 包括的なテストフロー
make test-qa

# リリース前の完全チェック
make test-release
```

#### **CI/CD環境での使用**
```bash
# CI用の短時間フロー
make test-ci

# QA環境での包括的テスト
make test-qa
```

### テスト実装の特徴

#### **🎯 TDD (Test-Driven Development)**
- t-wadaの手法に基づくテストファースト開発
- Red → Green → Refactor サイクルの実践

#### **🔄 動的ブランチ追従テスト**
- Dev側のブランチ変更を自動検出
- Ops側の自動ブランチ切り替えをテスト
- リアルタイムでの同期動作を検証

#### **⚡ 実際のGit操作テスト**
- 本物のGitリポジトリでの動作確認
- 実際のコミット、ブランチ操作を検証
- ファイルの追加、変更、削除をテスト

#### **🕒 継続的動作テスト**
- 30秒間の継続的開発シミュレーション
- 複数ファイルの自動作成・コミット
- 自動同期・fixup動作の検証

### テスト環境要件

#### **必須環境**
- Go 1.20+
- Git（ユニバーサル操作用）

#### **推奨環境**
- Windows 10/11（VHDX機能テスト用）
- 管理者権限（VHDX作成・マウント用）

### テスト結果の確認

```bash
# テストカバレッジの確認
make test-coverage
open coverage.html

# パッケージ別詳細カバレッジ
make test-coverage-detail
ls coverage-*.out

# セキュリティ・脆弱性チェック
make security-scan
make vuln-check
```

## アーキテクチャ

```
FixupCommitSyncManager/
├── cmd/                    # CLI コマンド実装
├── internal/
│   ├── config/            # 設定管理（ブランチ設定削除済み）
│   ├── sync/              # ファイル同期（動的ブランチ追従）
│   ├── fixup/             # Fixup コミット（動的ブランチ追従）
│   ├── vhdx/              # VHDX 管理
│   ├── logger/            # ログシステム
│   ├── retry/             # リトライ機能
│   ├── notify/            # 通知システム
│   └── utils/             # ユーティリティ
└── main.go                # エントリーポイント
```

## 動的ブランチ追従の仕組み

1. **ブランチ検出**: Dev リポジトリのカレントブランチを `git branch --show-current` で検出
2. **ブランチ切り替え**: Ops リポジトリを同じブランチに自動切り替え
3. **ブランチ作成**: 必要に応じてローカルまたはリモートから新規ブランチを作成
4. **差分検出**: Dev の直前コミット（HEAD^）との差分を検出
5. **同期実行**: 検出した差分を Ops リポジトリの同じブランチにコミット

## ライセンス

MIT License

## コントリビューション

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## サポート

- 📖 詳細仕様: [SPECIFICATION.md](SPECIFICATION.md)
- 🔧 実装ガイド: [CLAUDE.md](CLAUDE.md)
- 🐛 バグ報告: [Issues](https://github.com/your-org/FixupCommitSyncManager/issues)

---

**FixupCommitSyncManager** - Windows 環境でのソースコード同期を効率化