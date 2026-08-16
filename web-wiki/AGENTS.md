# web-wiki 前端知识库

## 概述

`web-wiki/` 是 Lumina 的 **Wiki Reader 独立前端**：只读 Wiki 阅读 SPA，用于渲染 RepoWiki 生成的文档站点。通过 `/api/v1/wiki/:wikiId/...` REST API 与后端通信，部署在 `/wiki/` 基础路径下。

## 技术栈

- **React 19** + **TanStack Router**（基于文件路由）
- **Tailwind CSS 4** + **shadcn/ui**（Radix 基础组件）
- **react-markdown** + **remark-math** + **rehype-katex** 渲染 Markdown 与数学公式
- **motion**（动画变体）
- **highlight.js** + **rehype-mermaid** 代码高亮与图表
- **Orama** 客户端搜索 + **@orama/tokenizers/mandarin** CJK 分词器
- **axios** 进行 API 请求，使用 Cookie 鉴权（无 Bearer Token）

## 目录结构

```text
web-wiki/
├── package.json            # pnpm 管理；端口 3001；依赖 @lumina/components workspace 包
├── tsconfig.json           # 路径别名 #/* / @/* → ./src/*
├── vite.config.ts          # Vite + TanStack Router 插件；Orama 独立 vendor chunk
├── src/
│   ├── main.tsx            # 应用入口
│   ├── router.tsx          # TanStack Router 配置
│   ├── routeTree.gen.ts    # 自动生成路由树（勿手动编辑）
│   ├── styles.css          # Tailwind 主题 + CSS 变量
│   ├── test-setup.ts       # Vitest 全局 setup（修复 Node≥25 预置 localStorage 遮蔽 jsdom）
│   ├── test-setup.test.ts  # setup 兜底实现测试
│   ├── routes/             # 基于文件的路由
│   │   ├── __root.tsx      # 根布局（头部导航）
│   │   └── wiki/           # /wiki/$wikiId / /wiki/$wikiId/$
│   │       ├── $wikiId.$.tsx              # catch-all 路由薄壳（懒加载 wiki-catchall-page）
│   │       └── $wikiId/
│   │           ├── index.tsx              # Wiki 入口路由薄壳（懒加载 wiki-index-page）
│   │           ├── wiki-index-page.tsx    # 入口页实现（manifest + home + buildPageTree 渲染卡片/正文）
│   │           └── wiki-catchall-page.tsx # catch-all 实现（按 _splat 读取指定页正文）
│   ├── components/         # 业务组件
│   │   ├── docs-page.tsx         # 三栏文档布局（替换旧 wiki-layout）
│   │   ├── page-tree-sidebar.tsx # 导航侧边栏（替换旧 wiki-sidebar）
│   │   ├── markdown-renderer.tsx # MarkdownRenderer（封装共享 Markdown + proseArticle 排版类）
│   │   ├── search.tsx            # Orama 客户端搜索（Cmd+K）
│   │   ├── toc.tsx               # 文章目录 + 滚动高亮
│   │   ├── breadcrumb.tsx        # 面包屑导航
│   │   ├── prev-next.tsx         # 上一篇/下一篇
│   │   ├── password-gate.tsx     # 密码门认证
│   │   └── password-input.tsx    # 密码输入
│   ├── hooks/              # React Hooks
│   │   └── useWikiAuth.ts  # Wiki 访问密码认证 Hook
│   └── lib/                # 工具与 API 客户端
│       ├── api-client.ts   # wikiApi + wikiReaderApi 类型定义
│       ├── source.ts       # PageTree 构建 + 图标映射 + 节点查找
│       ├── frontmatter.ts  # frontmatter 解析 + TOC 提取
│       └── utils.ts
```

## 约定

- **包管理器**：必须使用 `pnpm`；禁止 npm/yarn。
- **路径别名**：`#/*` 与 `@/*` 均映射到 `./src/*`；组件内优先使用 `#/`。
- **基础路径**：Wiki Reader 部署在 `/wiki/` 下；路由以 `/wiki/$wikiId` 为根。
- **认证方式**：Cookie 鉴权（`withCredentials: true`），无 Token 刷新逻辑。
- **API 基地址**：`wikiApi` 使用 `/api/v1`，所有接口封装在 `lib/api-client.ts`。
- **Markdown 安全**：react-markdown 默认不渲染 raw HTML（不使用 rehype-raw），无需额外 sanitize。
- **共享组件**：shadcn/ui 组件、markdown 渲染原语、motion 变体统一从 `@lumina/components` 共享包导入。
- **自动生成文件**：`routeTree.gen.ts` 由 TanStack Router 插件自动生成，禁止手动编辑。
- **状态管理**：使用 `useState`/`useEffect` 本地状态；数据获取使用 axios 直接请求。
- **内容格式**：`.mdx` 内容由后端解析 YAML frontmatter，前端直接消费 DTO 字段（title/description/icon/last_updated 等）。
- **manifest 路径**：导航中的 `path` 字段**不带扩展名**（如 `overview`、`modules/auth`）；磁盘文件实际为 `.mdx`。
- **Fenced 语法**：`:::callout`、`:::card`、`:::steps` 等 fenced 块由 `@lumina/components` 的 remark 插件渲染，不需要 rehype-raw。
- **侧边栏展开状态**：`page-tree-sidebar.tsx` 将展开状态持久化到 `localStorage`，键名为 `wiki-sidebar-expanded-{wikiId}`。
- **Markdown 排版**：正文渲染统一走 `markdown-renderer.tsx` 的 `MarkdownRenderer`，它封装共享 `@lumina/components/markdown` 并套 `proseArticle` 排版类；禁止绕过它直接使用 Markdown 组件（`proseArticle` 缺失会导致代码块黑底回归）。
- **路由懒加载**：`$wikiId/index.tsx` 与 `$wikiId.$.tsx` 是懒加载薄壳（`lazyRouteComponent`），实际实现分别在 `wiki-index-page.tsx` / `wiki-catchall-page.tsx`，用于首屏体积优化。
- **测试 setup**：`test-setup.ts` 作为 Vitest `setupFiles` 全局执行，修复 Node≥25 预置 `localStorage` 访问器遮蔽 jsdom 注入的问题（内存 `createMemoryStorage` 兜底），并 `afterEach` 自动 `cleanup()`。
- **BREAKING**：旧版 `.md` 页面不再兼容；后端与前端均只处理 `.mdx`，无 `.md` fallback。

## 调试路径

1. 路由 404 → 确认 `routes/wiki/` 文件路径与 `$wikiId`/`$` 参数匹配。
2. 401 鉴权失败 → 检查 Cookie 是否已设置，密码门是否正确提交。
3. 侧边栏为空 → 确认 `getManifest` 返回 `navigation` 数组，字段名与 `WikiNavItem` 对齐。
4. Markdown 不渲染 → 检查 `@lumina/components/markdown` 插件链（remark-gfm + remark-math + rehype-highlight + rehype-katex + rehype-mermaid + remark-fenced-blocks）。
5. Mermaid 图表不渲染 → 确认 `window.mermaid` 已加载，且 markdown 组件已初始化。
6. 搜索无结果 → 确认 Orama 索引已构建、`@orama/tokenizers/mandarin` 分词器已加载，且 `leaves` 非空。
7. 目录高亮异常 → 检查正文 heading 的 `id` 是否与 `extractToc` 生成的 slug 一致。
8. 切换页面时侧边栏重载 → 确认 `docs-page.tsx` 的 motion key 为 `wikiId`，而非 `currentPagePath`。
9. 代码块黑底 → 确认渲染经过 `markdown-renderer.tsx` 的 `proseArticle` 排版类（`markdown-renderer.test.tsx` 守卫）。
10. 测试环境 localStorage 报错 → 检查 `test-setup.ts` 是否正确注入 `vitest.config.ts` 的 `setupFiles`。

## 常用命令

```bash
cd web-wiki
pnpm install      # 安装依赖
pnpm dev          # 开发服务器（端口 3001）
pnpm build        # 类型检查 + 生产构建
pnpm test         # 运行 Vitest 测试
pnpm lint         # ESLint 检查
pnpm format       # Prettier 格式化 + ESLint 自动修复
pnpm check        # Prettier 格式检查
```
