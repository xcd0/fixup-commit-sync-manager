# ============================================================================
# FixupCommitSyncManager プロジェクト用 Makefile
# ============================================================================

# ============================================================================
# プロジェクト設定
# ============================================================================
# go.modからモジュール名とバイナリ名を自動取得。
MOD := $(shell go list -m 2>/dev/null || echo "unknown")
BIN := fixup-commit-sync-manager

# ディレクトリ設定。
BIN_DIR := ./bin
DIST_DIR := dist

# バージョン情報。
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION ?= 1.0.0
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

# ビルドフラグ設定。
FLAGS_VERSION := -X main.version=$(VERSION) -X main.revision=$(REVISION) -X main.buildDate=$(BUILD_DATE)
FLAG_DEBUG := -race -gcflags="-N -l"
FLAG_RELEASE := -a -tags netgo -trimpath -ldflags='-s -w -extldflags="-static" $(FLAGS_VERSION) -buildid='

# OS判定と実行ファイル拡張子設定。
UNAME_S := $(shell uname -s 2>/dev/null || echo "Windows")
ifeq ($(UNAME_S),Linux)
	OS := linux
	EXE_EXT :=
else
	OS := windows
	EXE_EXT := .exe
endif

# アーキテクチャ判定。
UNAME_M := $(shell uname -m 2>/dev/null || echo "x86_64")
ifeq ($(UNAME_M),x86_64)
	ARCH := amd64
else ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
else ifeq ($(UNAME_M),arm64)
	ARCH := arm64
else
	ARCH := amd64
endif

# ============================================================================
# メインターゲット
# ============================================================================
.PHONY: all help build clean release test lint fmt vet install deps update-deps
.PHONY: cross-compile run get-upx
.DEFAULT_GOAL := help

all: help

help: ## ヘルプを表示。
	@echo "利用可能なコマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# ビルドターゲット
# ============================================================================
build: ## デバッグビルド。
	@echo "デバッグビルドを実行中..."
	@mkdir -p $(BIN_DIR)
	go build $(FLAG_DEBUG) -o $(BIN_DIR)/$(BIN)$(EXE_EXT) .
	@echo "ビルド完了: $(BIN_DIR)/$(BIN)$(EXE_EXT)"

release-win: get-upx ## Windows リリースビルド + UPX圧縮。
	@echo "Windows リリースビルドを実行中..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(FLAG_RELEASE) -o $(BIN_DIR)/$(BIN).exe .
	@if [ -f "upx$(EXE_EXT)" ] || command -v upx >/dev/null 2>&1; then \
		echo "UPXで圧縮中..."; \
		if [ -f "upx$(EXE_EXT)" ]; then \
			./upx$(EXE_EXT) --lzma $(BIN_DIR)/$(BIN).exe || echo "UPX圧縮に失敗しましたが続行します。"; \
		else \
			upx --lzma $(BIN_DIR)/$(BIN).exe || echo "UPX圧縮に失敗しましたが続行します。"; \
		fi; \
	else \
		echo "UPXが見つかりません。圧縮をスキップします。"; \
	fi
	@echo "Windows リリースビルド完了: $(BIN_DIR)/$(BIN).exe"

release-linux: get-upx ## Linux リリースビルド + UPX圧縮。
	@echo "Linux リリースビルドを実行中..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(FLAG_RELEASE) -o $(BIN_DIR)/$(BIN) .
	@chmod +x $(BIN_DIR)/$(BIN)
	@if [ -f "upx$(EXE_EXT)" ] || command -v upx >/dev/null 2>&1; then \
		echo "UPXで圧縮中..."; \
		if [ -f "upx$(EXE_EXT)" ]; then \
			./upx$(EXE_EXT) --lzma $(BIN_DIR)/$(BIN) || echo "UPX圧縮に失敗しましたが続行します。"; \
		else \
			upx --lzma $(BIN_DIR)/$(BIN) || echo "UPX圧縮に失敗しましたが続行します。"; \
		fi; \
	else \
		echo "UPXが見つかりません。圧縮をスキップします。"; \
	fi
	@echo "Linux リリースビルド完了: $(BIN_DIR)/$(BIN)"

release: clean release-win release-linux ## 両OS用リリース一括ビルド。

# ============================================================================
# 開発用ターゲット
# ============================================================================
test: ## テスト実行。
	@echo "テストを実行中..."
	GOOS=windows go test -v ./...

test-coverage: ## テストカバレッジを計測。
	@echo "テストカバレッジを計測中..."
	GOOS=windows go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポート: coverage.html"

benchmark: ## ベンチマークテスト実行。
	@echo "ベンチマークテストを実行中..."
	GOOS=windows go test -bench=. -benchmem ./...

lint: ## リント実行。
	@echo "リントを実行中..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lintがインストールされていません。go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest でインストールしてください。"; \
	fi

fmt: ## コードフォーマット。
	@echo "コードをフォーマット中..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

vet: ## go vet実行。
	@echo "go vetを実行中..."
	go vet ./...

install: release-$(OS) ## バイナリをGOPATH/binにインストール。
	@echo "バイナリをインストール中..."
	go install $(FLAG_RELEASE)

run: build ## ビルドして実行。
	@echo "アプリケーションを実行中..."
	$(BIN_DIR)/$(BIN)$(EXE_EXT) --help

run-example: build ## 設定ファイル作成例を実行。
	@echo "設定ファイル作成例を実行中..."
	$(BIN_DIR)/$(BIN)$(EXE_EXT) init-config --dry-run

# ============================================================================
# 依存関係管理
# ============================================================================
deps: ## 依存関係を取得。
	@echo "依存関係を取得中..."
	go mod download
	go mod verify

update-deps: ## 依存関係を更新。
	@echo "依存関係を更新中..."
	go get -u ./...
	go mod tidy

vendor: ## vendorディレクトリを作成。
	@echo "vendorディレクトリを作成中..."
	go mod vendor

mod-tidy: ## go mod tidyを実行。
	@echo "go mod tidyを実行中..."
	go mod tidy

# ============================================================================
# クロスコンパイル
# ============================================================================
cross-compile: clean ## 複数プラットフォーム向けビルド。
	@echo "クロスコンパイルを実行中..."
	@mkdir -p $(DIST_DIR)
	@for os in windows linux; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then ext=".exe"; else ext=""; fi; \
			echo "Building for $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(FLAG_RELEASE) -o $(DIST_DIR)/$(BIN)-$$os-$$arch$$ext . || continue; \
		done; \
	done
	@echo "クロスコンパイル完了。$(DIST_DIR)ディレクトリを確認してください。"

# ============================================================================
# 掃除関連
# ============================================================================
clean-dist: ## distディレクトリを削除。
	@echo "distディレクトリを削除中..."
	@rm -rf $(DIST_DIR)

clean-vendor: ## vendorディレクトリを削除。
	@echo "vendorディレクトリを削除中..."
	@rm -rf vendor

clean-binary: ## バイナリファイルを削除。
	@echo "バイナリファイルを削除中..."
	@rm -rf $(BIN_DIR)

clean-coverage: ## カバレッジファイルを削除。
	@echo "カバレッジファイルを削除中..."
	@rm -f coverage.out coverage.html

clean: clean-binary clean-coverage ## 基本的な掃除。
	@echo "基本的な掃除完了。"

clean-all: clean clean-dist clean-vendor ## 全ての生成ファイルを削除。
	@echo "全ての掃除完了。"

# ============================================================================
# 開発サポート
# ============================================================================
watch: ## ファイル変更を監視してビルド（要: entr）。
	@if command -v entr >/dev/null 2>&1; then \
		echo "ファイル変更を監視中... (Ctrl+Cで終了)"; \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "entrコマンドが必要です。apt install entr または brew install entr でインストールしてください。"; \
	fi

debug: ## デバッガでビルド・実行（要: dlv）。
	@if command -v dlv >/dev/null 2>&1; then \
		echo "デバッガでビルド・実行中..."; \
		dlv debug; \
	else \
		echo "Delveデバッガが必要です。go install github.com/go-delve/delve/cmd/dlv@latest でインストールしてください。"; \
	fi

mod-graph: ## モジュール依存関係をグラフ表示。
	@echo "モジュール依存関係:"
	go mod graph

# ============================================================================
# 外部ツール取得
# ============================================================================
get-upx: ## UPXを取得（GitHub API経由）。
	@if [ ! -f "upx$(EXE_EXT)" ] && ! command -v upx >/dev/null 2>&1; then \
		echo "UPXをダウンロード中..."; \
		if [ "$(OS)" = "windows" ]; then \
			UPX_ASSET="win64.zip"; \
		else \
			UPX_ASSET="amd64_linux.tar.xz"; \
		fi; \
		UPX_URL=$$(curl -s https://api.github.com/repos/upx/upx/releases/latest \
			| grep -o "\"browser_download_url\":\"[^\"]*$$UPX_ASSET\"" \
			| cut -d'"' -f4); \
		curl -L "$$UPX_URL" -o upx_pkg; \
		if [ "$(OS)" = "windows" ]; then \
			unzip -jo upx_pkg "upx*/upx.exe"; \
		else \
			tar -xf upx_pkg --strip-components=1 "*/upx"; \
			chmod +x ./upx; \
		fi; \
		rm -f upx_pkg; \
		echo "UPXダウンロード完了。"; \
	else \
		echo "UPXは既にインストールされています。"; \
	fi

# ============================================================================
# セキュリティ関連
# ============================================================================
security-scan: ## セキュリティスキャン（要: gosec）。
	@if command -v gosec >/dev/null 2>&1; then \
		echo "セキュリティスキャンを実行中..."; \
		gosec ./...; \
	else \
		echo "gosecが必要です。go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest でインストールしてください。"; \
	fi

vuln-check: ## 脆弱性チェック（Go 1.18+）。
	@echo "脆弱性チェックを実行中..."
	@if go version | grep -q "go1.1[89]" || go version | grep -q "go1.[2-9][0-9]"; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	else \
		echo "Go 1.18以上が必要です。"; \
	fi

# ============================================================================
# 情報表示
# ============================================================================
info: ## プロジェクト情報を表示。
	@echo "=== プロジェクト情報 ==="
	@echo "モジュール: $(MOD)"
	@echo "バイナリ名: $(BIN)"
	@echo "バージョン: $(VERSION)"
	@echo "リビジョン: $(REVISION)"
	@echo "ビルド日時: $(BUILD_DATE)"
	@echo "OS: $(OS)"
	@echo "アーキテクチャ: $(ARCH)"
	@echo "Go バージョン: $$(go version 2>/dev/null || echo 'Go未インストール')"
	@echo "=== 環境情報 ==="
	@echo "GOPATH: $$(go env GOPATH 2>/dev/null || echo '未設定')"
	@echo "GOROOT: $$(go env GOROOT 2>/dev/null || echo '未設定')"
	@echo "GOOS: $$(go env GOOS 2>/dev/null || echo '未設定')"
	@echo "GOARCH: $$(go env GOARCH 2>/dev/null || echo '未設定')"

status: ## Gitステータスとプロジェクト状態を表示。
	@echo "=== Git ステータス ==="
	@git status --porcelain 2>/dev/null || echo "Gitリポジトリではありません"
	@echo ""
	@echo "=== ファイル状態 ==="
	@echo "go.mod: $$([ -f go.mod ] && echo '存在' || echo '未作成')"
	@echo "go.sum: $$([ -f go.sum ] && echo '存在' || echo '未作成')"
	@echo "$(BIN_DIR)/$(BIN)$(EXE_EXT): $$([ -f $(BIN_DIR)/$(BIN)$(EXE_EXT) ] && echo '存在' || echo '未作成')"
	@echo "main.go: $$([ -f main.go ] && echo '存在' || echo '未作成')"

# ============================================================================
# アプリケーション固有のコマンド
# ============================================================================
config-example: build ## 設定ファイル例を生成。
	@echo "設定ファイル例を生成中..."
	$(BIN_DIR)/$(BIN)$(EXE_EXT) init-config --config example-config.hjson

validate-example: config-example ## 例設定ファイルを検証。
	@echo "例設定ファイルを検証中..."
	$(BIN_DIR)/$(BIN)$(EXE_EXT) validate-config --config example-config.hjson --verbose

demo: build ## デモ実行（ヘルプとバージョン表示）。
	@echo "=== FixupCommitSyncManager デモ ==="
	@echo ""
	@echo "1. バージョン情報:"
	$(BIN_DIR)/$(BIN)$(EXE_EXT) --version
	@echo ""
	@echo "2. ヘルプ情報:"
	$(BIN_DIR)/$(BIN)$(EXE_EXT) --help
	@echo ""
	@echo "3. サブコマンド例 (sync --help):"
	$(BIN_DIR)/$(BIN)$(EXE_EXT) sync --help
