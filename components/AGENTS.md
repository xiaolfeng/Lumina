<!-- deep-init:synced@fa47a98 -->

# components 知识库

## 概述

`@lumina/components` 是 pnpm workspace 共享包：集中管理 shadcn/ui、Markdown 渲染原语、motion 动画变体和微明主题 CSS，被控制台（`web/`）与 Wiki Reader（`web-wiki/`）共同消费。独立于两个前端应用的构建配置，改这里会同时影响两端视觉与排版。

## 目录结构

```text
components/
├── package.json            # 包名 @lumina/components；exports 映射 ui/markdown/motion/theme
├── components.json         # shadcn/ui（new-york、zinc、lucide）
├── vitest.config.ts        # Vitest（markdown/remark-fenced-blocks 等）
├── eslint.config.js
├── tsconfig.json
└── src/
    ├── index.ts            # 包入口（再导出较少，消费方多走子路径 exports）
    ├── ui/                 # shadcn/ui 组件（accordion…tooltip，含 splitter）
    ├── markdown/           # Markdown 原语
    │   ├── markdown.tsx            # 完整渲染（GFM + 数学 + 高亮 + mermaid + fenced）
    │   ├── markdown-lite.tsx       # 轻量渲染
    │   ├── markdown-mermaid.tsx    # Mermaid 初始化
    │   ├── remark-fenced-blocks.ts # :::callout/card/steps 解析
    │   ├── fenced-components.tsx   # fenced 块 React 组件（含协议白名单防 XSS）
    │   ├── table-of-contents.tsx   # TOC 提取与渲染
    │   ├── prose.ts                # proseQuestion / proseHint / proseArticle
    │   └── index.ts
    ├── motion/             # 缓动函数与全局动画变体
    ├── lpw/                # LPW 1.1 文档渲染引擎（三类节点模型与出版级美学）
    │   ├── index.ts        # 导出入口（31 种组件注册 + 渲染器 + 解析器 + 契约表）
    │   ├── types.ts        # 1.1 节点模型、props 与契约类型定义
    │   ├── parser.ts       # 1.1 节点树 JSON 解析器
    │   ├── registry.ts     # kind:type 统一节点注册表
    │   ├── renderer.tsx    # 三类节点统一渲染器与层级/变体走查
    │   ├── contract.ts     # 策略组/变体契约表（与 Go 端快照对齐）
    │   ├── annotation.tsx  # 批注系统（装饰层徽标 + 波浪划线）
    │   ├── diagnostics.tsx # 结构化诊断卡与 props 敏感词遮盖
    │   ├── source-viewer.tsx # 相对资源基址计算与文件直渲外壳
    │   ├── document-viewer.tsx # 出版级文档展示壳
    │   ├── layout/         # 纯结构布局组件（7 种 pattern）
    │   ├── container/      # 受控美化外壳（section/panel/details/tabs）
    │   └── block/          # 26 种原子内容块与 ECharts/Mermaid 视口
    ├── styles/
    │   └── theme.css       # 静烛 v1 色盘与 CSS 变量
    ├── hooks/
    │   └── use-mobile.ts
    └── lib/
        └── utils.ts        # cn()
```

## 导航指南

| 任务 | 位置 | 说明 |
| --- | --- | --- |
| 新增 shadcn/ui 组件 | `src/ui/` | 在仓库根或 `components/` 执行 `pnpm dlx shadcn@latest add <name>`，输出到本包 |
| 改主题色盘 / 圆角 | `src/styles/theme.css` | 静烛 v1（`--sea-ink` / `--lagoon` / `--palm` / `--sand` / `--foam`），`--radius:0px` |
| 改 Markdown 插件链 | `src/markdown/markdown.tsx` | remark-gfm + remark-math + remark-fenced-blocks；rehype-highlight / katex / mermaid / slug |
| 改排版层级 | `src/markdown/prose.ts` | `proseQuestion`（Q&A 主文）/ `proseHint`（补充）/ `proseArticle`（Wiki 正文，防代码块黑底） |
| 新增 fenced 块 | `remark-fenced-blocks.ts` + `fenced-components.tsx` | 语法与组件必须成对出现 |
| 新增 LPW 块/布局/容器 | `src/lpw/` | 三类节点契约见 docs/engineering/design/0003；经 `@lumina/components/lpw` 导出 |
| 改动画变体 | `src/motion/` | 两端通过 `@lumina/components/motion` 导入 |
| 改 `cn()` | `src/lib/utils.ts` | 同时经 package exports `./utils` 暴露 |

## 约定

- **单一事实源**：ui / markdown / motion / 主题只在本包维护，因为 `web` 与 `web-wiki` 必须共用同一视觉语言；禁止在任一前端再拷一份。
- **包导出走 `exports`**：消费方按 `@lumina/components/ui/button`、`@lumina/components/markdown`、`@lumina/components/motion`、`@lumina/components/theme.css` 导入，避免从 `src/` 相对路径穿透。
- **shadcn 安装目标是本包**：`components.json` 指向 `src/ui`，CLI 加组件不会落到 `web/src`。
- **`proseArticle` 是 Wiki 正文的护栏**：它把 `<pre>` 设为白底墨字；绕过它直接用 `Markdown` 会让代码块回到黑底。
- **fenced 块不依赖 rehype-raw**：`:::callout` 等由 remark 插件转成组件，保持默认不渲染 raw HTML；Card 等链接采用严格协议白名单（http/https/mailto/相对路径）并拦截控制字符防范 XSS。

## 反模式

- 禁止在 `web/` 或 `web-wiki/` 内重新创建已迁入本包的 ui / markdown / motion / 主题文件——两端会立刻分叉。
- 禁止在 `web/` 或 `web-wiki/` 内实现 LPW 渲染——必须从 `@lumina/components/lpw` 导入。
- 禁止把组件级样式写进本包 `theme.css`；主题文件只放设计 token。
- 禁止手动在 `web/src` 下新建 shadcn 文件来「图省事」；必须走 CLI 加到本包。

## 调试路径

1. 前端找不到组件 → 核对 `package.json` 的 `exports` 与导入路径是否走子路径而非包根。
2. 代码块黑底 → 确认渲染套了 `proseArticle`（Wiki）或对应 `proseQuestion`/`proseHint`。
3. fenced 块不生效 → 检查 `remark-fenced-blocks` 是否进入 markdown 插件链，且 `fenced-components` 已注册对应类型。
4. 主题色不对 → 确认应用入口导入了 `@lumina/components/theme.css`，而不是只改了前端本地 `styles.css`。
5. 测试失败 → 在 `components/` 下跑 `pnpm test`（Vitest 覆盖 remark-fenced-blocks 等）。
