# Lumina Q&A 14 种题型全景概览与快速选型指南

Lumina Q&A 系统提供 14 种专为软件研发与 Agent 决策设计的富交互题型，按功能分为 **选择类**、**输入类**、**展示类** 和 **评分类** 四大家族。

---

## 🧭 快速选型决策树

```text
需要与用户交互？
 │
 ├── ① 让用户做选择
 │    ├── 仅选择方案/环境 (单选) ─────────────▶ [select]
 │    ├── 需要详细对比各方案优缺点 (pros/cons) ─▶ [options]
 │    └── 勾选多个功能模块/组件 ───────────────▶ [multi-select]
 │
 ├── ② 需要用户提供内容/输入
 │    ├── 简单二选一确认 (是/否) ───────────────▶ [boolean]
 │    ├── 自由文本/多行描述 ───────────────────▶ [text]
 │    ├── 代码片段/正则表达式 (带高亮编辑器) ──▶ [code]
 │    ├── 架构图/设计截图 (带 OTP 下载) ───────▶ [image]
 │    └── 配置文件/数据附件 (带 OTP 下载) ───────▶ [file]
 │
 ├── ③ 需要用户审阅与审批
 │    ├── 代码修改前后对比 (Before/After) ────▶ [diff]
 │    ├── 实施方案/迁移步骤分段审批 ───────────▶ [plan]
 │    └── 架构/文档内容逐段评审 ───────────────▶ [review]
 │
 └── ④ 需要量化评估与反馈
      ├── 连续数值/百分比调节 ────────────────▶ [slider]
      ├── 任务/需求优先级排列 ────────────────▶ [rank]
      └── 多维度独立打分 (1-5星) ─────────────▶ [rate]
```

---

## 📊 14 种题型核心特性速查表

| 族系 | 题型 (`question_type`) | 必须参数 | 是否支持 Supplement | 核心返回标记 | 详情文档 |
|---|---|---|---|---|---|
| **选择类** | `select` | `options` | ✅ 支持 (强烈推荐) | `[ANSWER]`, `[OPTION_DESCRIPTION]`, `[SUPPLEMENT]` | [`select.md`](./select.md) |
| | `multi-select` | `options` | ✅ 支持 | `[ANSWER]`, 多项 `---` 分隔, `[OPTION]` | [`multi-select.md`](./multi-select.md) |
| | `options` | `options` (含 `pros`/`cons`) | ✅ 支持 | `[ANSWER]`, `[DESCRIPTION]`, `[SUPPLEMENT]` (用户 feedback) | [`options.md`](./options.md) |
| **输入类** | `text` | 基础三字段 | ❌ | `[ANSWER]` 文本正文 | [`text.md`](./text.md) |
| | `boolean` | 基础三字段 | ❌ | `[ANSWER] 是` 或 `[ANSWER] 否` | [`boolean.md`](./boolean.md) |
| | `code` | 基础三字段 (`config.language`) | ❌ | `[ANSWER]` 代码正文, `[LANGUAGE]` | [`code.md`](./code.md) |
| | `image` | 基础三字段 | ❌ | `[FILE_NAME]`, `[DOWNLOAD_PATH]`, `[DOWNLOAD_URL]` | [`image.md`](./image.md) |
| | `file` | 基础三字段 (`config.accept`) | ❌ | `[FILE_NAME]`, `[DOWNLOAD_PATH]`, `[DOWNLOAD_URL]` | [`file.md`](./file.md) |
| **展示类** | `diff` | `config.before`, `config.after` | ❌ | `[ANSWER]`, `[FINAL]` (最终代码), `[FEEDBACK]` | [`diff.md`](./diff.md) |
| | `plan` | `config.sections` | ❌ | `[ANSWER]`, `[PLAN_DETAIL]`, `[REVISIONS]`, `[FEEDBACK]` | [`plan.md`](./plan.md) |
| | `review` | `config.sections` | ❌ | `[ANSWER]`, `[FEEDBACK]` | [`review.md`](./review.md) |
| **评分类** | `slider` | 基础三字段 (`config.min/max`) | ❌ | `[ANSWER]` 数值 | [`slider.md`](./slider.md) |
| | `rank` | `options` | ❌ | `[ANSWER] 1. A → 2. B → 3. C` | [`rank.md`](./rank.md) |
| | `rate` | `options` (每项独立打分) | ❌ | `[ANSWER] 项1: 分数, 项2: 分数` | [`rate.md`](./rate.md) |

*注：基础三字段指的是 `session_id`、`question_type`、`content`。*
