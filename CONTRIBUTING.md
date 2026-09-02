# 贡献指南

感谢你愿意给 Lumina 提补丁。本仓库是 MIT 许可的公开项目；接受外部 Pull Request，但合并权在维护者。行为底线见 [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)。

架构边界（模块互不调用、分层、MCP 认证）以 [ARCHITECTURE.md](ARCHITECTURE.md) 为准。AI 编码助手应先读 [AGENTS.md](AGENTS.md)。

## 环境

- Go 1.25.11（见 `go.mod`）
- pnpm（根目录 `pnpm-workspace.yaml` 管理 `web` / `web-wiki` / `components`）
- PostgreSQL、Redis
- 可选：`swag`、`golangci-lint`、`gh`（发版）

```bash
cp .env.example .env
pnpm install
go mod tidy
```

`.env` 已被 gitignore，不要提交。生产务必设置 `XLF_INITIALIZE_TOKEN`、`LLM_ENCRYPT_SECRET`、`REPOWIKI_HMAC_SECRET`。

## 本地开发

```bash
make dev-backend        # 双前端打包 + Swagger + 跑后端（推荐联调）
make dev                # 只跑后端（跳过前端构建）
make dev-frontend       # 控制台 Vite，端口 3000
make dev-wiki-frontend  # Wiki Reader Vite，端口 3001
```

验证：

```bash
curl http://localhost:8080/api/v1/health/ping
```

Swagger UI 仅在 `XLF_DEBUG=true` 时注册。

一键生产构建：`make generate`（或 `make build`）——前端产物写入 `resources/*/dist`，再编译 Go。

## 质量门

```bash
make fmt
make test
make vet
make lint          # 未安装 golangci-lint 时跳过
```

前端（在对应包或仓库根）：

```bash
pnpm --filter @lumina/components test
# 或 cd web && pnpm test / cd web-wiki && pnpm test
```

提交前请保证你改动的层有对应测试或可复现的手工验证说明。涉及 UI 的改动要写清你验证过哪些路由，而不是只贴一张截图。

## 分支与提交

- 从最新 `master` 拉功能分支，PR 也打回 `master`。
- 提交信息用中文、说清「做了什么 / 为什么」，scope 与仓库现有风格对齐（如 `feat(接入):`、`fix(交互):`）。
- 不要添加 `Co-Authored-By`；作者就是提交者本人。
- 不要 `git push` 到维护者没请你推的远端分支以外的地方——PR 从你的 fork 来即可。

## Pull Request

PR 描述里写：动机、行为变化、验证方式、是否破坏 API / MCP 工具名 / 存储格式。

维护者会核对：

- 有没有跨层调用或五模块互调
- 有没有把密钥写进清单、日志或测试夹具
- `docs/swagger.json` / `docs/swagger.yaml` / `docs/docs.go` 是否被手改（这些只能 `make swag` 生成）
- 新增实体是否实现 `GetGene()` 并加入 `WithAutoMigrate`

## 编码约定（摘要）

完整条文在 `AGENTS.md` 与各子目录 `AGENTS.md`。最少记住：

- 后端：`route` → `handler` → `logic` → `repository`；共享能力放 `internal/service`。
- 使用 `bamboo-base-go` 前查官方文档，禁止凭记忆调用。
- 环境变量走 `xEnv.GetEnv*`；Info 键走 `internal/constant/info_key.go`。
- DTO 按业务域分包、按操作拆文件（`api/<domain>/create.go` 等）。
- 前端包管理只用 pnpm；shadcn 组件加到 `@lumina/components`，不要在 `web/src` 再造一份。
- MCP 工具名保持 snake_case，与 `internal/mcp` 现网一致。

## 许可证

贡献默认按仓库 [MIT License](LICENSE) 授权。
