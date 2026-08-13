# 变量定义，方便后续维护
MAIN_FILE = main.go
SWAG_CMD = swag
SWAG_FLAGS = --parseDependency
OUTPUT_BIN = lumina

# Docker / Release（GitHub Actions 驱动）
REPO = xiaolfeng/Lumina
DOCKER_WORKFLOW = docker-publish.yml
WATCH ?= # 置 1 时触发后阻塞观察，如 make publish VERSION=v0.1.0 WATCH=1

.DEFAULT_GOAL := help

.PHONY: help install swag run dev dev-backend dev-frontend dev-wiki-frontend build-frontend build-wiki-frontend tidy fmt test vet lint build generate
.PHONY: validate-version docker-build publish watch

# 显示帮助信息
help:
	@echo "Lumina · 微明 - 可用命令"
	@echo ""
	@echo "初始化命令:"
	@echo "  make install             - 安装根 pnpm workspace 依赖"
	@echo ""
	@echo "开发命令:"
	@echo "  make swag                - 生成 Swagger 文档"
	@echo "  make run                 - 运行后端程序"
	@echo "  make dev                 - 生成文档并运行后端 (跳过前端构建)"
	@echo "  make dev-backend         - 一键构建并运行后端 (包含前端构建)"
	@echo "  make dev-frontend        - 运行主前端开发服务器"
	@echo "  make dev-wiki-frontend   - 运行 Wiki Reader 前端开发服务器"
	@echo ""
	@echo "构建命令:"
	@echo "  make generate            - 一键构建：主前端 → Wiki Reader → Swagger → Go 编译（运行前请先 make install 确保依赖最新）"
	@echo "  make build               - 同 generate"
	@echo "  make build-frontend      - 仅构建主前端 (产出 resources/web/dist)"
	@echo "  make build-wiki-frontend - 仅构建 Wiki Reader 前端 (产出 resources/web-wiki/dist)"
	@echo ""
	@echo "质量命令:"
	@echo "  make tidy                - 整理 Go 模块"
	@echo "  make fmt                 - 格式化 Go 代码"
	@echo "  make test                - 运行 Go 测试"
	@echo "  make vet                 - 运行 go vet 静态检查"
	@echo "  make lint                - 运行 golangci-lint (未安装则跳过)"
	@echo ""
	@echo "Docker / Release（GitHub Actions 驱动，需 gh CLI 已登录）:"
	@echo "  make docker-build VERSION=vX.X.X  - 触发 CI 构建镜像并推送 Docker Hub"
	@echo "  make publish VERSION=vX.X.X      - 触发 CI 构建并创建 GitHub Release"
	@echo "                                     (版本含 -beta/-alpha 等后缀 → prerelease，镜像打 beta-latest)"
	@echo "  make watch               - 观察最近一次 workflow run 输出"
	@echo ""
	@echo "示例:"
	@echo "  make dev"
	@echo "  make dev-backend"
	@echo "  make build"
	@echo "  make docker-build VERSION=v0.1.0"
	@echo "  make publish VERSION=v0.1.1-beta.1"
	@echo ""

# 安装所有 workspace 依赖（根 pnpm workspace）
install:
	pnpm install

# 提取出的 Swagger 生成目标
swag:
	$(SWAG_CMD) init -g $(MAIN_FILE) $(SWAG_FLAGS)

# 提取出的运行目标
run:
	chmod +x $(OUTPUT_BIN) && ./$(OUTPUT_BIN)

tidy:
	go mod tidy

# 格式化 Go 代码
fmt:
	go fmt ./...

# 运行 Go 测试
test:
	go test ./...

# 组合目标：先生成文档，再运行后端程序
dev-backend: generate run

# 后端开发：生成 Swagger 文档后运行（跳过前端构建）
dev: swag run

# 前端开发服务器
dev-frontend:
	cd web && pnpm dev

# Wiki Reader 前端开发服务器
dev-wiki-frontend:
	cd web-wiki && pnpm dev

# 一键构建：前端打包 → Wiki Reader 前端打包 → Swagger 文档 → Go 编译
generate: build-frontend build-wiki-frontend swag
	go build -o $(OUTPUT_BIN)

# build 是 generate 的别名
build: generate

# 仅构建主前端（产出 resources/web/dist）
build-frontend:
	cd web && pnpm install && pnpm build

# 仅构建 Wiki Reader 前端（产出 web-wiki/dist）
build-wiki-frontend:
	cd web-wiki && pnpm install && pnpm build

# 静态检查
vet:
	go vet ./...

# 代码 lint（若 golangci-lint 未安装则优雅跳过）
lint:
	@which golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed, skipping"; exit 0; }
	@golangci-lint run ./...

# 校验版本号：vX.Y.Z 或 vX.Y.Z-预发布后缀
validate-version:
	@test -n "$(VERSION)" || { echo "ERROR: 请指定 VERSION=vX.X.X（如 make publish VERSION=v0.1.0）"; exit 1; }
	@echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$$' \
		|| { echo "ERROR: 非法版本号 '$(VERSION)'，示例: v0.1.0 / v0.1.1-beta.1"; exit 1; }

# 触发 GitHub Actions 构建镜像并推送 Docker Hub（不创建 Release）
docker-build: validate-version
	@echo "触发 GitHub Actions 构建镜像 $(VERSION) ..."
	@gh workflow run $(DOCKER_WORKFLOW) -R $(REPO) -f version=$(VERSION) -f publish=false
	@if [ "$(WATCH)" = "1" ]; then gh run watch -R $(REPO); fi

# 触发 GitHub Actions 构建镜像并发布 GitHub Release（版本含 - 后缀自动识别为 prerelease）
publish: validate-version
	@echo "触发 GitHub Actions 构建并发布 $(VERSION) ..."
	@gh workflow run $(DOCKER_WORKFLOW) -R $(REPO) -f version=$(VERSION) -f publish=true
	@if [ "$(WATCH)" = "1" ]; then gh run watch -R $(REPO); fi

# 观察最近一次 workflow run 的实时输出
watch:
	gh run watch -R $(REPO)
