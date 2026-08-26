# Q&A 题型规范：options（差异化对比专用）

## 📌 定位与适用场景
- **用途**：展示带优缺点（`pros` / `cons`）对比的选项卡片，让用户在充分知情下做权衡抉择。
- **适用场景**：核心技术选型（如 DB 选型、框架选型）、架构重构方案对比、算法权衡等。
- **重要约束**：**每个 option 必须包含 `pros`（优点列表）和 `cons`（缺点列表）**！缺失会导致后端拒绝。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "options",
  "content": "string (必填，方案对比提示，支持 Markdown)",
  "supplement": true,
  "options": [
    {
      "label": "string (必填，方案名称)",
      "description": "string (必填，方案核心概述)",
      "pros": ["string (必填，优点 1)", "string (必填，优点 2)"],
      "cons": ["string (必填，缺点 1)", "string (必填，缺点 2)"]
    }
  ]
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "options",
  "content": "## 状态同步方案技术选型\n\n请权衡以下两套实时状态同步方案并做出选择：",
  "supplement": true,
  "options": [
    {
      "label": "WebSocket 全双工通道",
      "description": "基于已有的 Hub 基础设施建立双向长连接",
      "pros": [
        "实时性极高（毫秒级延迟）",
        "复用现有连接，开销低",
        "支持服务端主动推送与多设备协同"
      ],
      "cons": [
        "需维护心跳与重连机制",
        "有状态服务，水平扩容复杂度稍高"
      ]
    },
    {
      "label": "短轮询 (HTTP Polling)",
      "description": "前端定时请求 REST API 获取最新状态",
      "pros": [
        "实现极其简单，无状态",
        "天然兼容所有网络代理与防火墙"
      ],
      "cons": [
        "存在轮询周期延迟",
        "空轮询会产生无意义的 QPS 浪费"
      ]
    }
  ]
}
```

---

## 📥 后端返回格式

```text
[ANSWER] WebSocket 全双工通道
[DESCRIPTION] 基于已有的 Hub 基础设施建立双向长连接
[SUPPLEMENT] （用户在前端提交的额外 feedback 或补充说明）
```

---

## 💡 注意事项与避坑指南
1. **强制 pros/cons**：如果只是普通单选，请用 `select`；只要用了 `options`，就必须认真列举每种方案的优缺点（利弊分析）。
2. **用户 Feedback**：返回的 `[SUPPLEMENT]` 会携带用户在 UI 上填写的选择原因或附加说明。
