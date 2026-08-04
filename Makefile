# 项目 Makefile
# Go 工具链：优先使用环境变量 GO/GOROOT 指定，默认使用 PATH 中的 go。
# 工具链版本由 go.mod 的 toolchain 指令锁定（go1.26.5）。
# 示例：make GO=/path/to/go GOROOT=/path/to/go

GO ?= go
GOENV := PATH="$(dir $(shell command -v $(GO) 2>/dev/null || echo $(GO)))$$PATH"

APP_NAME = lum-server
BUILD_DIR = ./build
MAIN_PATH = .

.PHONY: all build run test lint fmt clean docker tools

all: lint test build

# ---- 构建 ----
build:
	$(GOENV) $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOENV) $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_PATH)

build-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOENV) $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-mac $(MAIN_PATH)

# ---- 运行 ----
run:
	$(GOENV) $(GO) run $(MAIN_PATH)

# ---- 测试 ----
test:
	$(GOENV) $(GO) test -v -race -count=1 ./...

test-coverage:
	$(GOENV) $(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOENV) $(GO) tool cover -html=coverage.out -o coverage.html

# ---- 代码检查 ----
lint:
	$(GOENV) $(GO) vet ./...

fmt:
	$(GOENV) $(GO) fmt ./...

# ---- 清理 ----
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# ---- Docker ----
docker:
	docker build -t $(APP_NAME):latest -f Dockerfile ..

# ---- 工具安装 ----
tools:
	$(GOENV) $(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GOENV) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
