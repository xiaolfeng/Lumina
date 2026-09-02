# 求助渠道

把问题送到对的地方，Issue 区才能保持可检索。

## 优先级

1. **安全漏洞** → [SECURITY.md](SECURITY.md)。不要开公开 Issue。
2. **行为准则相关** → [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) 里的私密联系方式。
3. **用法 / 接入 / 部署** → 先读 README、控制台「接入指南」页、[ARCHITECTURE.md](ARCHITECTURE.md) 与 [CONTRIBUTING.md](CONTRIBUTING.md)。仍不清楚再开 GitHub Issue。
4. **缺陷与功能请求** → [GitHub Issues](https://github.com/xiaolfeng/Lumina/issues)。一份 Issue 只谈一件事；写清版本（git tag 或 commit）、环境（自托管 / Docker）、复现步骤和期望行为。
5. **准备改代码** → 先看 [CONTRIBUTING.md](CONTRIBUTING.md)，再发 PR。

没有官方论坛、没有承诺时限的即时通讯支持。维护者是个人，回复可能隔几天。

## 不要开 Issue 的情况

- 密钥、`.env`、OAuth 令牌、Webhook HMAC 泄露——先轮换密钥，安全细节走 SECURITY.md。
- 作业式「请帮我从零搭一套」且没有具体报错。
- 与 Lumina 无关的竹框架（`bamboo-base-go`）问题——去对应仓库。
