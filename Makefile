# 变量定义，方便后续维护
MAIN_FILE = main.go
SWAG_CMD = swag
SWAG_FLAGS = --parseDependency
BUILD_SCRIPT = script/build-docker.sh
SCRIPT_DIR = script

.DEFAULT_GOAL := help

.PHONY: help swag run dev

# 显示帮助信息
help:
	@echo "BambooBase - 可用命令"
	@echo ""
	@echo "开发命令:"
	@echo "  make swag       - 生成 Swagger 文档"
	@echo "  make run        - 运行程序"
	@echo "  make dev        - 生成文档并运行 (推荐)"
	@echo ""
	@echo "示例:"
	@echo "  make dev"
	@echo ""

# 提取出的 Swagger 生成目标
swag:
	$(SWAG_CMD) init -g $(MAIN_FILE) $(SWAG_FLAGS)

# 提取出的运行目标
run:
	go run $(MAIN_FILE)

tidy:
	go mod tidy

# 组合目标：先生成文档，再运行程序
# 以后你只需要执行 `make dev` 就可以一键起飞了！
dev: swag run
