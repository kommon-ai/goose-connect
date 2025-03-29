.PHONY: test lint clean

# デフォルトのターゲット
all: test lint

# テストの実行
test:
	@echo "Running tests..."
	go test -v ./...

# リントの実行
lint:
	@echo "Running linters..."
	golangci-lint run

# クリーンアップ
clean:
	@echo "Cleaning up..."
	go clean
	rm -f goose-connect

# 開発用の依存関係のインストール
deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# ビルド
build:
	@echo "Building..."
	go build -o goose-connect

# インストール
install:
	@echo "Installing..."
	go install

# ヘルプ
help:
	@echo "Available targets:"
	@echo "  all      - Run all checks (test and lint)"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linters"
	@echo "  clean    - Clean up build artifacts"
	@echo "  deps     - Install development dependencies"
	@echo "  build    - Build the binary"
	@echo "  install  - Install the binary"
	@echo "  help     - Show this help message" 