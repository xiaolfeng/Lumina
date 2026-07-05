# web-wiki 前端知识库

## 概述

`web-wiki/` 是 Lumina 的 **Wiki Reader 独立前端**：只读 Wiki 阅读 SPA，用于渲染 RepoWiki 生成的文档站点。通过 `/api/v1/wiki/:wikiId/...` REST API 与后端通信，部署在 `/wiki/` 基础路径下。

## 技术栈

- **React 19** + **TanStack Router**（基于文件路由）
- **Tailwind CSS 4** + **shadcn/ui**（Radix 基础组件）
- **react-markdown** + **rehype-sanitize** 渲染 Markdown
- **highlight.js** + **rehype-mermaid** 代码高亮与图表
- **axios** 进行 API 请求，使用 Cookie 鉴权（无 Bearer Token）

## 目录结构

```text
web-wiki/
├── package.json            # pnpm 管理；端口 3001
├── tsconfig.json           # 路径别名 #/* / @/* → ./src/*
├── vite.config.ts          # Vite + TanStack Router 插件
├── src/
│   ├── main.tsx            # 应用入口
│   ├── router.tsx          # TanStack Router 配置
│   ├── routeTree.gen.ts    # 自动生成路由树（勿手动编辑）
│   ├── styles.css          # Tailwind 主题 + CSS 变量
│   ├── routes/             # 基于文件的路由
│   │   ├── __root.tsx      # 根布局（头部导航）
│   │   └── wiki/           # /wiki/$wikiId / /wiki/$wikiId/$
│   ├── components/         # 业务组件
│   │   ├── wiki-layout.tsx     # Wiki 页面布局
│   │   ├── wiki-sidebar.tsx    # 导航侧边栏
│   │   ├── markdown-renderer.tsx
│   │   ├── password-gate.tsx
│   │   └── password-input.tsx
│   ├── hooks/              # React Hooks
│   │   ├── useWiki.ts
│   │   └── useWikiAuth.ts
│   └── lib/                # 工具与 API 客户端
│       ├── api-client.ts   # wikiApi + wikiReaderApi
│       └── utils.ts
```

## 约定

- **包管理器**：必须使用 `pnpm`；禁止 npm/yarn。
- **路径别名**：`#/*` 与 `@/*` 均映射到 `./src/*`；组件内优先使用 `#/`。
- **基础路径**：Wiki Reader 部署在 `/wiki/` 下；路由以 `/wiki/$wikiId` 为根。
- **认证方式**：Cookie 鉴权（`withCredentials: true`），无 Token 刷新逻辑。
- **API 基地址**：`wikiApi` 使用 `/api/v1`，所有接口封装在 `lib/api-client.ts`。
- **Markdown 安全**：统一通过 `rehype-sanitize` 处理；禁止 `dangerouslySetInnerHTML`。
- **自动生成文件**：`routeTree.gen.ts` 由 TanStack Router 插件自动生成，禁止手动编辑。
- **状态管理**：使用 `useState`/`useEffect` 本地状态；数据获取使用 axios 直接请求。

## 调试路径

1. 路由 404 → 确认 `routes/wiki/` 文件路径与 `$wikiId`/`$` 参数匹配。
2. 401 鉴权失败 → 检查 Cookie 是否已设置，密码门是否正确提交。
3. 侧边栏为空 → 确认 `getManifest` 返回 `navigation` 数组，字段名与 `WikiNavItem` 对齐。
4. Markdown 不渲染 → 检查 `react-markdown` 插件链顺序，`rehype-sanitize` 必须在最前。
5. Mermaid 图表不渲染 → 确认 `window.mermaid` 已加载，且 `markdown-renderer.tsx` 已初始化。

## 常用命令

```bash
cd web-wiki
pnpm install      # 安装依赖
pnpm dev          # 开发服务器（端口 3001）
pnpm build        # 类型检查 + 生产构建
pnpm lint         # ESLint 检查
pnpm format       # Prettier 格式化 + ESLint 自动修复
```
