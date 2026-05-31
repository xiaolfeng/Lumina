# 变量定义，方便后续维护
MAIN_FILE = main.go
SWAG_CMD = swag
SWAG_FLAGS = --parseDependency
BUILD_SCRIPT = script/build-docker.sh
SCRIPT_DIR = script

.DEFAULT_GOAL := help

.PHONY: help swag run dev-backend dev-frontend tidy fmt test

# 显示帮助信息
help:
	@echo "BambooBase - 可用命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make swag         - 生成 Swagger 文档"
	@echo "  make run          - 运行后端程序"
	@echo "  make dev-backend  - 生成文档并运行后端 (推荐)"
	@echo "  make dev-frontend - 运行前端开发服务器"
	@echo ""
	@echo "质量命令:"
	@echo "  make tidy         - 整理 Go 模块"
	@echo "  make fmt          - 格式化代码"
	@echo "  make test         - 运行测试"
	@echo ""
	@echo "示例:"
	@echo "  make dev-backend"
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
