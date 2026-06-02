# 变量定义，方便后续维护
MAIN_FILE = main.go
SWAG_CMD = swag
SWAG_FLAGS = --parseDependency
BUILD_SCRIPT = script/build-docker.sh
SCRIPT_DIR = script
OUTPUT_BIN = lumina

.DEFAULT_GOAL := help

.PHONY: help swag run dev-backend dev-frontend tidy fmt test build generate

# 显示帮助信息
help:
	@echo "Lumina · 微明 - 可用命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make swag         - 生成 Swagger 文档"
	@echo "  make run          - 运行后端程序"
	@echo "  make dev-backend  - 生成文档并运行后端 (推荐)"
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
	@echo ""
	@echo "示例:"
	@echo "  make dev-backend"
	@echo "  make build"
	@echo ""

# 提取出的 Swagger 生成目标
swag:
	$(SWAG_CMD) init -g $(MAIN_FILE) $(SWAG_FLAGS)

# 提取出的运行目标
run:
	go run $(MAIN_FILE)

tidy:
	go mod tidy

# 组合目标：先生成文档，再运行后端程序
dev-backend: swag run

# 前端开发服务器
dev-frontend:
	cd web && pnpm dev

# 一键构建：前端打包 → Swagger 文档 → Go 编译
generate: build-frontend swag
	go build -o $(OUTPUT_BIN) . && chmod +x $(OUTPUT_BIN) && ./$(OUTPUT_BIN)

# build 是 generate 的别名
build: generate

# 仅构建前端（产出 web/dist）
build-frontend:
	cd web && pnpm install && pnpm build
