# 安全策略

## 支持范围

Lumina 目前处于 `v1.0.0-beta.*` 预发布。安全修复只打在 **`master` 上的最新标签**（当前基线为 `v1.0.0-beta.18`）。更早的 beta 标签不提供补丁回溯。尚未声明长期支持版本。

自托管实例请尽快跟到最新标签或 `master` 头。镜像由 tag 触发的 GitHub Actions 发布。

## 上报漏洞

**不要**用公开 Issue、PR 或讨论区报告可被利用的漏洞。

请使用 GitHub 私密公告：

https://github.com/xiaolfeng/Lumina/security/advisories/new

报告里尽量包含：

- 受影响版本或 commit
- 复现步骤（PoC 以最小化为限，不要附带可直接打生产的利用链）
- 实际影响（未授权访问、密钥泄露、路径穿越、OAuth 绕过等）
- 你是否已经在公开场合讨论过

如果 GitHub 私密公告不可用，再通过维护者 GitHub 账号 [@xiaolfeng](https://github.com/xiaolfeng) 私信，标题标明「Security」。

## 响应时限

这是个人维护的仓库，以下是目标而不是 SLA：

| 阶段 | 目标 |
| --- | --- |
| 确认收到 | 3 个自然日内 |
| 初步定性（可利用 / 误报 / 重复） | 7 个自然日内 |
| 修复或公开说明 | 视严重程度，争取 30 个自然日内 |

请给维护者留出静默期，在约定披露日之前不要公开细节。

## 特别敏感的面

下列区域出问题通常视为高优先级：

- MCP 认证（`McpAuth`：OAuth `lum_at_` 与 API Key 回退）
- OAuth 2.1 授权码 / PKCE / 令牌存储（摘要而非原文）
- WebAuthn Origin / RP ID 推导与 `XLF_BIOMETRIC_ALLOWED_ORIGINS`
- Wiki / Preview 的路径与 HTML 沙盒
- Webhook HMAC
- `LLM_ENCRYPT_SECRET` 保护的 API Key 与 SSH 私钥
- 一次性初始化令牌 `XLF_INITIALIZE_TOKEN`

安全相关约定的架构表述见 [ARCHITECTURE.md](ARCHITECTURE.md)。
