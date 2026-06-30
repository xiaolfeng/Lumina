# 变量定义，方便后续维护
MAIN_FILE = main.go
SWAG_CMD = swag
SWAG_FLAGS = --parseDependency
BUILD_SCRIPT = script/build-docker.sh
SCRIPT_DIR = script
OUTPUT_BIN = lumina

.DEFAULT_GOAL := help

.PHONY: help swag run dev dev-backend dev-frontend tidy fmt test vet lint build generate

# 显示帮助信息
help:
	@echo "Lumina · 微明 - 可用命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make swag         - 生成 Swagger 文档"
	@echo "  make run          - 运行后端程序"
	@echo "  make dev          - 生成文档并运行后端 (跳过前端构建)"
	@echo "  make dev-backend  - 一键构建并运行后端 (包含前端构建)"
	@echo "  make dev-frontend - 运行前端开发服务器"
	@echo ""
	@echo "构建命令:"
	@echo "  make generate     - 一键构建：前端打包 → Swagger → Go 编译"
	@echo "  make build        - 同 generate"
	@echo ""
	@echo "质量命令:"
	@echo "  make tidy         - 整理 Go 模块"
	@echo "  make fmt          - 格式化代码"
	@echo "  make test         - 运行测试"
	@echo "  make vet          - 运行 go vet 静态检查"
	@echo "  make lint         - 运行 golangci-lint (未安装则跳过)"
	@echo ""
	@echo "示例:"
	@echo "  make dev"
	@echo "  make dev-backend"
	@echo "  make build"
	@echo ""

# 提取出的 Swagger 生成目标
swag:
	$(SWAG_CMD) init -g $(MAIN_FILE) $(SWAG_FLAGS)

# 提取出的运行目标
run:
	chmod +x $(OUTPUT_BIN) && ./$(OUTPUT_BIN)

tidy:
	go mod tidy

# 组合目标：先生成文档，再运行后端程序
dev-backend: generate run

# 后端开发：生成 Swagger 文档后运行（跳过前端构建）
dev: swag run

# 前端开发服务器
dev-frontend:
	cd web && pnpm dev

# 一键构建：前端打包 → Swagger 文档 → Go 编译
generate: build-frontend swag
	go build -o $(OUTPUT_BIN)

# build 是 generate 的别名
build: generate

# 仅构建前端（产出 web/dist）
build-frontend:
	cd web && pnpm install && pnpm build

# 静态检查
vet:
	go vet ./...

# 代码 lint（若 golangci-lint 未安装则优雅跳过）
lint:
	@which golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed, skipping"; exit 0; }
	@golangci-lint run ./...
