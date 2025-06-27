## 1. プログラム概要

- **名称**：FixupCommitSyncManager
- **目的**：
  - Windows 環境で Dev リポジトリ⇔Ops リポジトリ間のソース差分を軽量に自動同期し、定期的に自動 fixup コミットを行う
  - VHDX を用いた隔離初期化機能、設定生成／検証機能を含む総合運用プラットフォームを提供

## 2. 用語定義

| 用語        | 意味                                    |
| --------- | ------------------------------------- |
| Dev リポジトリ | 開発者が日常的に編集・ビルドを行うローカル Git リポジトリ       |
| Ops リポジトリ | 差分同期および履歴整理用に用意されたローカル Git リポジトリ      |
| VHDX      | Windows 上の仮想ディスクファイル                  |
| サブコマンド    | CLI で指定する個別機能（例：sync、fixup、init-vhdx） |
| ロックファイル   | 同期一時停止を指示するファイル（デフォルト：`.sync-paused`） |

## 3. 外部インターフェース

### 3.1 CLI 構成

```bash
FixupCommitSyncManager [--config <path>] <subcommand> [--dry-run] [--verbose]
```

- `<subcommand>`：下表の子コマンド名
- 共通オプション：
  - `--config <path>`：設定ファイルパス (デフォルト `config.hjson`)
  - `--dry-run`      ：実際のファイル操作やコミットを行わず、ログのみ出力
  - `--verbose`      ：標準出力にも詳細ログを出力

| サブコマンド            | 機能概要                                                  |
| ----------------- | ----------------------------------------------------- |
| `init-config`     | 対話型ウィザードで HJSON 設定ファイルの雛形を生成                          |
| `validate-config` | 設定ファイルの構文チェックおよび必須項目／バージョン互換性検証                       |
| `init-vhdx`       | VHDX ファイル作成→マウント→Ops リポジトリ初期 clone→リモート URL 置換→アンマウント |
| `mount-vhdx`      | 指定 VHDX をマウント                                         |
| `unmount-vhdx`    | 指定 VHDX をアンマウント                                       |
| `snapshot-vhdx`   | VHDX のスナップショット作成・一覧・ロールバック                            |
| `sync`            | Dev→Ops 間でソース差分（tracked+新規 .cpp/.h/.hpp）の同期＆自動コミット    |
| `fixup`           | Ops リポジトリで定期的に `--fixup` + `--autosquash` コミットを実行     |
| `help`            | サブコマンド一覧およびヘルプ表示                                      |

### 3.2 設定ファイル設定項目

設定ファイル (`config.hjson`) で指定可能な項目を示します。

| 設定項目               | 説明                                      | 型例                                    | 既定値                                   |
| ------------------ | --------------------------------------- | ------------------------------------- | ------------------------------------- |
| devRepoPath        | Dev リポジトリのローカルパス（必須）                    | `"C:\\path\\to\\dev-repo"`            | ―                                     |
| opsRepoPath        | Ops リポジトリのローカルパス（必須）                    | `"C:\\path\\to\\ops-repo"`            | ―                                     |
| includeExtensions  | 同期対象とするファイル拡張子リスト（tracked 変更＋新規追加）      | `[".cpp", ".h", ".hpp"]`              | `[".cpp", ".h", ".hpp"]`              |
| includePatterns    | 同期対象に含める追加パスパターン（Glob 形式）               | `["src/**/*.cpp"]`                    | `[]`                                  |
| excludePatterns    | 同期対象から除外するパスパターン（Glob 形式）               | `["bin/**", "obj/**"]`                | `[]`                                  |
| syncInterval       | 差分同期モード実行間隔                             | `"5m"`                                | `"5m"`                                |
| pauseLockFile      | 同期一時停止用ロックファイル名                         | `".sync-paused"`                      | `".sync-paused"`                      |
| gitExecutable      | 実行する git コマンドパス                         | `"git"`                               | `"git"`                               |
| commitTemplate     | 同期コミット時のメッセージ雛形（テンプレート文字列）              | `"Auto-sync: ${timestamp} @ ${hash}"` | `"Auto-sync: ${timestamp} @ ${hash}"` |
| authorName         | 同期コミット時の著者名                             | `"Sync Bot"`                          | Git global 設定                         |
| authorEmail        | 同期コミット時の著者メール                           | `"sync-bot@example.com"`              | Git global 設定                         |
| fixupInterval      | 定期 fixup コミット実行間隔                       | `"1h"`                                | `"1h"`                                |
| fixupMessagePrefix | fixup コミット時のメッセージ接頭辞                    | `"fixup! "`                           | `"fixup! "`                           |
| autosquashEnabled  | `--autosquash` フラグ有効化                   | `true`                                | `true`                                |
| targetBranch       | Ops リポジトリの同期先ブランチ名                      | `"sync-branch"`                       | `"sync-branch"`                       |
| baseBranch         | fixup 対象のベースブランチ名                       | `"main"`                              | `"main"`                              |
| maxRetries         | Git/I/O 操作失敗時の最大リトライ回数                  | `3`                                   | `3`                                   |
| retryDelay         | リトライ間隔                                  | `"30s"`                               | `"30s"`                               |
| logLevel           | ログ出力レベル (`DEBUG`/`INFO`/`WARN`/`ERROR`) | `"INFO"`                              | `"INFO"`                              |
| logFilePath        | ログファイル出力パス                              | `"C:\\logs\\sync.log"`                | `"./sync.log"`                        |
| notifyOnError      | エラー時通知設定（例：Slack Webhook URL）           | `{ slackWebhookUrl: "https://..." }`  | ―                                     |
| dryRun             | 実際の操作を行わずログのみ出力                         | `false`                               | `false`                               |
| verbose            | 標準出力への詳細ログ出力                            | `false`                               | `false`                               |
| vhdxPath           | VHDX ファイルの作成先パス (init-vhdx 時必須)         | `"C:\\vhdx\\ops.vhdx"`                | ―                                     |
| vhdxSize           | VHDX ファイルサイズ                            | `"10GB"`                              | `"10GB"`                              |
| mountPoint         | VHDX マウント先ドライブ／パス (init-vhdx 時必須)       | `"X:"`                                | ―                                     |
| encryptionEnabled  | VHDX 暗号化を有効化                            | `true`                                | `false`                               |

## 4. 機能要件

### 4.1 init-config

1. 対話型プロンプトで必須項目入力
2. コメント付き HJSON テンプレート出力

### 4.2 validate-config

1. HJSON 構文チェック
2. 必須項目／型／バージョン互換性検証

### 4.3 init-vhdx

1. VHDX 作成・フォーマット
2. マウント → `git clone --single-branch --local`
3. `git remote set-url origin` に置換 → アンマウント

### 4.4 mount-vhdx / unmount-vhdx / snapshot-vhdx

- mount-vhdx: 指定 VHDX マウント
- unmount-vhdx: 指定 VHDX アンマウント
- snapshot-vhdx: スナップショット作成／一覧／ロールバック

### 4.5 sync

1. ロックファイル存在時スキップ
2. Dev 側で変更 tracked + 新規ソース検出
3. Ops へディレクトリ構造保持コピー／削除反映
4. `git add -u` → `git commit -m commitTemplate`

### 4.6 fixup

1. `git add -u`
2. `git commit --fixup=<baseBranch>@{now}` + `--autosquash`
3. ログ記録／リトライ＆通知

## 5. 非機能要件

- **環境**：Windows 10/11, Go 1.20+
- **性能**：1,000ファイル同期を5分以内
- **信頼性**：リトライ＆通知機能
- **拡張性**：サブコマンド追加容易

## 6. エラーハンドリング

- 例外捕捉 → リトライ → 通知
- Ctrl+C 時クリーンアップ

## 7. ロギング・通知

- レベル設定 (DEBUG/INFO/WARN/ERROR)
- ファイル & 標準出力
- 初期化完了サマリー通知

## 8. セキュリティ・運用性

- VHDX 暗号化オプション (BitLocker)
- 設定ファイルパーミッションチェック
- 対話型ウィザード & バリデーション

## 9. 配布形態

- Windows 用単一実行ファイル（.exe）
- 任意ディレクトリに配置後即実行可能

## 10. 実装順序

1. init-config
2. validate-config
3. init-vhdx
4. VHDX 管理(mount/unmount/snapshot)
5. sync
6. fixup
7. 共通ユーティリティ
8. 統合テスト

---

以上を参照して実装を進めてください。

