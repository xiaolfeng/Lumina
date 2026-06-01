# WEB 前端知识库

## 概述

TanStack Start（React 19）+ Tailwind CSS 4 + shadcn/ui 构建的 Lumina 前端应用，通过 REST API + SSE 与后端通信。

## 目录结构

```text
web/
├── package.json                # pnpm 管理；React 19 + TanStack Start + Tailwind CSS 4
├── vite.config.ts              # Vite 插件链：devtools → tailwind → tanstackStart → react
├── tsconfig.json               # TypeScript strict；路径别名 #/* → ./src/*
├── components.json             # shadcn/ui（new-york、zinc、lucide）
├── eslint.config.js            # @tanstack/eslint-config
├── prettier.config.js          # semi:false、singleQuote:true、trailingComma:all
├── public/                     # 静态资源（favicon、logo、manifest）
└── src/
    ├── router.tsx              # TanStack Router 入口（intent 预加载 + 滚动恢复）
    ├── routeTree.gen.ts        # 自动生成的路由树（请勿手动编辑）
    ├── styles.css              # 全局 CSS 变量 + Tailwind 主题 + 基础约束
    ├── routes/                 # 基于文件的路由
    │   ├── __root.tsx          # 根布局（HTML shell + DevTools）
    │   ├── _public.tsx         # 公开页面布局（Navbar + Outlet + Footer）
    │   ├── _auth.tsx           # 认证页面布局（品牌面板 + 表单区域）
    │   ├── _public/index.tsx   # 首页
    │   └── _auth/              # 认证页面
    │       ├── login.tsx       # 登录页
    │       └── reset-password.tsx  # 重置密码页
    ├── components/             # 组件
    │   ├── Navbar.tsx          # 全局导航栏
    │   ├── Footer.tsx          # 全局页脚
    │   └── ui/                 # shadcn/ui 组件（通过 CLI 添加）
    └── lib/
        └── utils.ts            # cn() 工具（clsx + tailwind-merge）
```

## 导航指南

| 任务                | 位置                     | 说明                                     |
| ------------------- | ------------------------ | ---------------------------------------- |
| 新增页面            | `src/routes/`            | 文件路径即路由路径；布局路由以 `_` 前缀  |
| 新增布局路由        | `src/routes/_<name>.tsx` | 含 `Outlet` 的布局组件                   |
| 新增通用组件        | `src/components/`        | Navbar/Footer 级别的全局组件             |
| 新增 shadcn/ui 组件 | `src/components/ui/`     | 通过 `pnpm dlx shadcn@latest add <name>` |
| 修改全局主题色      | `src/styles.css`         | 仅修改 CSS 变量和 `@theme inline`        |
| 修改路由配置        | `src/router.tsx`         | 预加载策略、滚动恢复等                   |
| 工具函数            | `src/lib/`               | 通用工具（如 `cn()`）                    |

## 约定

- **包管理器为 pnpm**：禁止使用 npm 或 yarn，确保 lock 文件一致性。
- **代码风格**：Prettier（`semi: false`、`singleQuote: true`、`trailingComma: "all"`）+ ESLint（`@tanstack/eslint-config`）。
- **路径别名**：`#/*` 映射到 `./src/*`；组件内使用 `#/components/xxx` 导入。
- **路由模式**：TanStack Start file-router；文件名即路由路径，`_` 前缀为布局路由。
- **shadcn/ui 管理**：组件通过 `pnpm dlx shadcn@latest add <component>` 添加，禁止手动创建 `ui/` 下的文件。
- **CSS 架构**：`styles.css` 仅负责 CSS 变量定义、`@theme inline` 映射、body 基础样式、`@layer base` 全局约束、`prefers-reduced-motion` 降级、全局排版工具类（`.display-title`、`.page-wrap`）。组件级样式由组件自身通过 Tailwind 类或 `<style>` 管理。
- **动画库**：使用 `motion`（Framer Motion 的轻量版）实现动画；变体（variants）定义在布局文件中并导出供子路由使用。
- **主题配色**：微明色盘（烛光暖褐系），亮/暗模式通过 `:root` / `.dark` CSS 变量切换；shadcn/ui 变量已兼容主题。

## 反模式

- 禁止在 `styles.css` 中编写组件级或页面级样式（如 `a {}`、`code {}`、`.island-shell` 等）。
- 禁止手动编辑 `routeTree.gen.ts`，它由 TanStack Router 插件自动生成。
- 禁止手动创建 `src/components/ui/` 下的 shadcn/ui 组件文件。
- 禁止使用 npm 或 yarn 安装依赖。
- 禁止在 `styles.css` 中引入第三方 CSS 库的完整样式。

## 调试路径

1. 页面 404 → 检查 `src/routes/` 文件路径是否正确匹配路由。
2. 样式异常 → 检查 `styles.css` 的 CSS 变量是否被覆盖；确认组件使用 Tailwind 类而非自定义 CSS。
3. shadcn/ui 组件不显示 → 确认通过 CLI 正确安装，检查 `components.json` 别名配置。
4. 路由跳转失败 → 检查 `router.tsx` 配置和 `routeTree.gen.ts` 是否为最新。
5. 动画不播放 → 确认 `motion` 导入是否正确（`motion/react`）；检查 `prefers-reduced-motion` 设置。
