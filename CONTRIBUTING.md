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
make check         # 版本规则测试 + gofmt + go vet + go test -race，与 CI 同一道门
make fmt
make test
make vet
make lint          # 未安装 golangci-lint 时跳过
```

发起 Pull Request 后，GitHub Actions 会自动跑 `make check`。镜像打包只在版本 tag 或 `make docker-build` / `make publish` 时执行，而且必须先通过同一道校验。

前端（在对应包或仓库根）：

```bash
pnpm --filter @lumina/components test
# 或 cd web && pnpm test / cd web-wiki && pnpm test
```

提交前请保证你改动的层有对应测试或可复现的手工验证说明。涉及 UI 的改动要写清你验证过哪些路由，而不是只贴一张截图。

## 发布

版本 tag 或 `make publish VERSION=vX.Y.Z`（亦可写作 `make public`）会先推送当前分支代码、在本地打版本 tag 并推送到远端。远端流水线会先校验版本连续性，再执行打包校验与发布。正式版本的 major / minor / patch 每次只能有一段递增一级；预发布版本序号必须连续，例如 `v1.1.0-beta.20` 后只能发布 `v1.1.0-beta.21` 或晋升 `v1.1.0`。跨版本预检失败时会清理本次 tag，并跳过全部构建。

通过预检后执行：

1. 构建并推送 Docker 镜像。
2. 通过 OpenAI 兼容的 `/chat/completions` 生成 Release 正文。
3. 使用 GoReleaser 构建 Linux、Windows、macOS、FreeBSD、NetBSD、OpenBSD 的 `amd64` / `arm64` 二进制。
4. 将各平台的原始二进制与 `checksums.txt` 直接上传到 GitHub Release。

AI Release 正文固定使用 `glm-5.3-flash` 和 `https://ai-intl.x-lf.com/v1`，仓库只需配置 Secret `AI_API_KEY`。模型不可用时自动使用 commit 清单。GoReleaser 配置位于 `.goreleaser.yaml`。校验、镜像构建或二进制发布失败时，工作流会删除本次提交对应的远端版本 tag；tag 已被移动到其他提交时会拒绝删除并明确报错。

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
