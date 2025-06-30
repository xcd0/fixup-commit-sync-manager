# テストデータディレクトリ

このディレクトリにはテスト実行時に使用するデータファイルが含まれています。

## ディレクトリ構造

```
test/
├── data/           # 固定テストデータ（git管理対象）
│   ├── test-config.hjson    # テスト用設定ファイル
│   └── README.md           # このファイル
├── repos/          # テスト用リポジトリ（実行時生成、git管理対象外）
├── vhdx/           # VHDX関連ファイル（実行時生成、git管理対象外）
├── config/         # 設定ファイルテスト用
├── temp/           # 一時ファイル（実行時生成、git管理対象外）
└── .gitignore      # テスト用gitignore
```

## テスト実行方法

### Windows環境でのテスト実行
```bash
GOOS=windows go test ./...
```

### 特定パッケージのテスト
```bash
GOOS=windows go test ./internal/vhdx
```

### 詳細出力付きテスト
```bash
GOOS=windows go test -v ./...
```

## 注意事項

- VHDXテストはWindows環境または管理者権限が必要です
- テスト実行前に不要なファイルを削除したい場合は `make clean-test` を実行してください
- 一部のテストは実際のVHDX操作を行うため時間がかかる場合があります