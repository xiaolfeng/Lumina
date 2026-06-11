# Q&A 交互式问答模块

## 概述

Q&A 模块是 Lumina 中 AI Agent 与用户进行富交互式问答的通道。Agent 通过 MCP 工具向用户推送结构化问题（选择、文本、评分等 14 种类型），用户通过浏览器实时接收并回答，回答即时返回给 Agent 继续处理。

## 交互流程

```
Agent (MCP)                    Lumina Server                     User (Browser)
    │                              │                                  │
    │── qa_session_create ────────>│                                  │
    │<── session_id + link ────────│                                  │
    │                              │                                  │
    │── qa_push_question ─────────>│── SSE/WebSocket 推送 ──────────>│
    │<── question_id ──────────────│                                  │
    │                              │                                  │
    │── qa_get_answer (阻塞) ─────>│     用户查看问题并回答             │
    │        ...等待...            │<── 提交回答 ─────────────────────│
    │                              │                                  │
    │<── [ANSWERED] answer ────────│                                  │
    │                              │                                  │
    │── qa_push_question ─────────>│── SSE/WebSocket 推送 ──────────>│
    │  (下一批问题...)             │                                  │
    └                              └                                  └
```

## MCP 工具

| 工具名 | 功能 | 说明 |
|--------|------|------|
| `qa_session_create` | 创建问答会话 | 支持 temporary(48h) 和 permanent 两种类型，需关联 project_id |
| `qa_session_get` | 查看会话状态 | 返回会话元数据和问题列表 |
| `qa_session_delete` | 删除会话 | 软删除，数据不可恢复 |
| `qa_push_question` | 推送问题 | 支持 14 种交互类型，返回 question_id 和 option_id 映射 |
| `qa_push_supplement` | 推送补充内容 | Markdown/HTML 格式，右侧面板渲染 |
| `qa_get_answer` | 阻塞等待回答 | 挂起直到用户回答 |
| `qa_reget_answer` | 批量获取回答 | 支持多个 question_id，返回 ANSWERED/PENDING/ERROR 状态 |
| `qa_what_question` | 查询问题类型 | 查看可用类型和参数说明 |

## 问题类型

### 选择类
- **single_choice**: 单选（从选项中选一个）
- **multi_choice**: 多选（从选项中选多个）
- **dropdown**: 下拉选择
- **cascading**: 级联选择

### 文本类
- **text_input**: 文本输入（单行/多行）
- **rich_text**: 富文本输入（Markdown）
- **code_input**: 代码输入（语法高亮）
- **textarea**: 多行文本区域

### 确认类
- **confirm**: 确认（是/否）
- **agree_terms**: 条款同意

### 评分类
- **rating**: 星级评分
- **slider**: 滑块评分

### 媒体类
- **file_upload**: 文件上传
- **image_upload**: 图片上传

## 会话管理

- 每个 Session 可关联一个 Project（project_id）
- temporary 类型 48 小时自动过期
- permanent 类型永不过期，支持跨设备
- 在线设备数自动追踪

## 通信机制

Lumina 使用 SSE（Server-Sent Events）和 WebSocket 实现问题即时推送：
- 新问题推送时，在线设备即时收到通知
- 回答提交后，Agent 端通过 MCP 消费回答队列
- 补充内容推送后，用户浏览器右侧面板即时渲染

## 数据库表

- `qa_session`: 会话表（标题、Agent、类型、状态、关联项目）
- `qa_question`: 问题表（类型、标题、选项、配置、状态、回答）
- `qa_supplement`: 补充内容表（目标类型、目标ID、内容格式、内容）

## 配置

- `QA_SESSION_MAX_DURATION`: Session 最大存活时间（秒，默认 604800 = 7 天）
