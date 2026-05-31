# RepoWiki 详细设计

## 核心流程

```
Git URL 输入
    │
    ▼
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Git Clone │────▶│ 文件扫描  │────▶│ 依赖图   │
│ (go-git)  │     │ + 过滤   │     │ + PageRank│
└──────────┘     └──────────┘     └──────────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │ LLM 分阶段分析 │
                                  │              │
                                  │ Pass 1: 概览  │
                                  │ Pass 2: 模块  │
                                  │ Pass 3: 架构  │
                                  │ Pass 4: 指南  │
                                  └──────┬───────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │ 文档组装      │
                                  │ Markdown 输出 │
                                  │ Mermaid 图表  │
                                  │ 元数据 JSON   │
                                  └──────┬───────┘
                                         │
                              ┌──────────┴──────────┐
                              ▼                     ▼
                       PostgreSQL 元数据       文件系统 Markdown
                       (项目、版本、状态)       (.lumina/repowiki/)
```

## 阶段一：Git 克隆与文件扫描

### Git 克隆

- 使用 go-git 库克隆目标仓库到本地临时目录
- 支持公开仓库和私有仓库（通过配置 credentials）
- 克隆完成后记录当前 commit hash 用于增量更新

### 文件扫描

- 遍历仓库目录树，收集文件列表
- 过滤规则：
  - 排除 `.git`、`node_modules`、`vendor`、`build` 等目录
  - 排除二进制文件、图片、字体等非文本内容
  - 尊重 `.gitignore` 规则
- 识别编程语言（根据文件扩展名统计）
- 检测入口文件（`main.go`、`index.ts`、`app.py` 等）

## 阶段二：依赖图与 PageRank

### 依赖图构建

- 解析 import / require 语句，建立文件间依赖关系
- 构建有向图（文件 → 文件、模块 → 模块）
- 识别核心模块（被依赖最多的节点）

### PageRank 排序

- 对文件按重要性排序（基于依赖图的 PageRank 算法）
- 高分文件优先送入 LLM 分析
- 确保在 token 限制内覆盖最重要的代码

## 阶段三：LLM 分阶段分析

LLM 通过 Agent SDK 调用，分四个阶段逐步深入：

### Pass 1：项目概览

- 项目定位与核心功能
- 技术栈识别（语言、框架、工具）
- 目录结构概述
- 入口点分析

### Pass 2：模块分析

- 模块清单与职责划分
- 模块间依赖关系图（Mermaid）
- 各模块关键文件与核心函数
- 模块接口定义

### Pass 3：架构设计

- 整体架构模式（分层、微服务、插件等）
- 架构图（Mermaid graph）
- 数据流图（Mermaid sequenceDiagram）
- 设计决策与约束

### Pass 4：阅读指南

- 基于 PageRank 的推荐阅读顺序
- 新人上手路径
- 关键代码段索引
- 常见问题与解答

## 阶段四：文档组装与输出

### 目录结构

```
.lumina/repowiki/{project_id}/
├── zh/
│   ├── 主页.md                    # Wiki 首页，项目概览与导航
│   ├── content/
│   │   ├── 项目概览.md            # Pass 1 产出
│   │   ├── 模块分析.md            # Pass 2 产出
│   │   ├── 架构设计.md            # Pass 3 产出
│   │   └── 阅读指南.md            # Pass 4 产出
│   └── meta/
│       └── repowiki-metadata.json # 侧边栏导航配置
└── en/
    └── ...                        # 英文版本（可选）
```

### 元数据 JSON 结构（参考值）

```json
{
  "navigation": [
    {
      "title": "项目概览",
      "path": "content/项目概览.md",
      "order": 1
    },
    {
      "title": "模块分析",
      "path": "content/模块分析.md",
      "order": 2
    },
    {
      "title": "架构设计",
      "path": "content/架构设计.md",
      "order": 3
    },
    {
      "title": "阅读指南",
      "path": "content/阅读指南.md",
      "order": 4
    }
  ],
  "home": "主页.md",
  "language": "zh",
  "project_name": "项目名称",
  "generated_at": "2026-05-31T00:00:00Z",
  "commit_hash": "abc1234"
}
```

## 增量更新策略

### 触发条件

- Agent 主动调用 `repoWiki_update`
- 对比当前 HEAD commit hash 与上次分析的 commit hash
- 如果无变更，跳过分析直接返回现有 Wiki

### 更新流程

1. 获取当前 HEAD commit hash
2. 与上次记录的 commit hash 对比
3. 计算变更文件列表（`git diff --name-only`）
4. 分析变更文件的依赖链（影响范围）
5. 仅对变更影响范围内的模块重新执行 LLM 分析
6. 未变更部分保留原 Wiki 内容
7. 更新 commit hash 记录

## 数据库模型（逻辑模型）

### 项目表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| name | string | 项目名称 |
| git_url | string | Git 仓库地址 |
| local_path | string | 本地克隆路径 |
| commit_hash | string | 当前分析的 commit |
| status | enum | 分析状态（pending / analyzing / completed / failed） |
| language | string | Wiki 语言（zh / en） |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### Wiki 版本表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| project_id | 雪花 ID | 关联项目 |
| commit_hash | string | 对应的 commit |
| llm_model | string | 使用的 LLM 模型 |
| file_count | int | 分析的文件数 |
| duration_ms | int | 分析耗时（毫秒） |
| created_at | timestamp | 生成时间 |
