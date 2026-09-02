# 变更记录

本文件遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号尽量靠近 [SemVer](https://semver.org/lang/zh-CN/)。预发布标签形如 `v1.0.0-beta.N`（含 `-` 的版本在 CI 里会标成 prerelease）。

本文件自 **v1.0.0-beta.18** 起作为变更的单一书面源。更早的 `v1.0.0-beta.2` 至 `v1.0.0-beta.17` 未在此逐条展开，请查阅 Git 标签与 GitHub Releases。

## [Unreleased]

### 文档

- 新增根位置 `ARCHITECTURE.md`（物理架构与边界）
- 新增 `CONTRIBUTING.md`、`CODE_OF_CONDUCT.md`、`SECURITY.md`、`SUPPORT.md`

## [1.0.0-beta.18] - 2026-09-02

当前预发布基线（与 `master` 头 `d1ef58c` 一致）。

相对此前若干 beta，本阶段产品已包含（按主题归并，不是该 tag 单次 commit 的完整 diff）：

### 新增

- MCP OAuth 2.1（授权服务器 + 资源服务器，PKCE 公共客户端，动态注册，consent 页）
- AI 插件运行时打包与市场清单（含 ZCode 专用 `marketplace.zcode.json`、Codex 仓库市场）
- 控制台接入指南：插件安装 / 手动 MCP / 技能安装 / 工具速览
- Preview 按文件类型分发：HTML/SVG 沙盒、Markdown、代码高亮
- Interact 会话进度条；历史问答按事件时间倒序置顶
- 控制台侧边栏用户弹出菜单（资料 / 退出）

### 修复

- WebAuthn 线上注册 RP ID 误回退 localhost
- ZCode 安装插件时 archive 源不被支持
- Interact 补充等待遮罩不统一、历史最新条目未置顶
- 控制台 Tabs 纵向滚动、表格横向溢出

### 变更

- MCP 端点认证改为 OAuth 优先、API Key 回退（`McpAuth`）

[Unreleased]: https://github.com/xiaolfeng/Lumina/compare/v1.0.0-beta.18...HEAD
[1.0.0-beta.18]: https://github.com/xiaolfeng/Lumina/releases/tag/v1.0.0-beta.18
