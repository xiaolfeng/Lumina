# WEB 前端知识库

## 概述

TanStack Start（React 19）+ Tailwind CSS 4 + shadcn/ui 构建的 Lumina 前端应用，通过 REST API + WebSocket 与后端通信。前端构建产物通过 `go:embed` 嵌入 Go 二进制，支持单文件部署。

## 目录结构

```text
web/
├── package.json                # pnpm 管理；React 19 + TanStack Start + Tailwind CSS 4
├── vite.config.ts              # Vite 插件链：devtools → tailwind → tanstackStart → react
├── tsconfig.json               # TypeScript strict；路径别名 #/* → ./src/*
├── components.json             # shadcn/ui（new-york、zinc、lucide）
├── eslint.config.js            # @tanstack/eslint-config
├── prettier.config.js          # semi:false、singleQuote:true、trailingComma:all
├── index.html                  # SPA 入口 HTML
├── public/                     # 静态资源（favicon、logo、manifest）
└── src/
    ├── main.tsx                # 应用入口（挂载 RouterProvider）
    ├── router.tsx              # TanStack Router 入口（intent 预加载 + 滚动恢复）
    ├── routeTree.gen.ts        # 自动生成的路由树（请勿手动编辑）
    ├── styles.css              # 全局 CSS 变量 + Tailwind 主题 + 基础约束
    ├── routes/                 # 基于文件的路由
    │   ├── __root.tsx          # 根布局（QueryClientProvider + 认证守卫 + DevTools）
    │   ├── _public.tsx         # 公开页面布局（Navbar + Outlet + Footer）
    │   ├── _public/
    │   │   ├── index.tsx       # 首页
    │   │   └── start.tsx       # 初始化向导页
    │   ├── auth.tsx            # 认证页面布局（品牌面板 + 表单区域）
    │   ├── auth/
    │   │   ├── login.tsx       # 登录页
    │   │   ├── new.tsx         # 首次初始化页（设置密码）
    │   │   └── reset-password.tsx  # 重置密码页
    │   ├── console.tsx         # 控制台布局（Sidebar + Breadcrumb + 认证守卫）
    │   ├── console/
    │   │   ├── index.tsx       # 控制台入口（重定向到 dashboard）
    │   │   ├── dashboard.tsx   # 仪表盘
    │   │   ├── apikey.tsx      # API Key 管理
    │   │   ├── project.tsx     # 项目管理
    │   │   ├── qa.tsx          # Q&A 会话管理（状态筛选 + 分页列表 + 删除）
    │   │   ├── qa/$sessionId.tsx  # Q&A 会话详情（问题列表）
    │   │   └── settings.tsx    # 系统设置
    │   └── interact.tsx        # Interact 交互布局（品牌栏 + 三栏主体）
    │   └── interact/
    │       └── index.tsx       # Interact 交互主页（WebSocket 连接 + 问题展示）
    ├── components/             # 组件
    │   ├── Navbar.tsx          # 公开页面导航栏
    │   ├── Footer.tsx          # 公开页面页脚
    │   ├── app-sidebar.tsx     # 控制台侧边栏（导航菜单）
    │   ├── data-table.tsx      # 通用数据表格组件
    │   ├── data-table-pagination.tsx # 通用分页组件
    │   ├── apikey/             # API Key 业务组件
    │   │   ├── columns.tsx     # 表格列定义
    │   │   ├── create-dialog.tsx   # 创建对话框
    │   │   ├── edit-dialog.tsx     # 编辑对话框
    │   │   ├── delete-dialog.tsx   # 删除确认对话框
    │   │   └── reset-dialog.tsx    # 重置确认对话框
    │   ├── project/            # 项目业务组件
    │   │   ├── columns.tsx     # 表格列定义
    │   │   ├── create-dialog.tsx   # 创建对话框
    │   │   ├── edit-dialog.tsx     # 编辑对话框
    │   │   └── delete-dialog.tsx   # 删除确认对话框
    │   ├── qa/                 # Q&A 管理业务组件
    │   │   ├── columns.tsx     # 会话列表列定义
    │   │   ├── delete-dialog.tsx   # 删除确认对话框
    │   │   ├── question-card.tsx   # 问题卡片展示
    │   │   └── session-detail.tsx  # 会话详情组件
    │   ├── interact/           # Interact 交互组件
    │   │   ├── types.ts        # 类型定义（Question/Session/SupplementItem）
    │   │   ├── session-sidebar.tsx # 会话侧边栏
    │   │   ├── question-card.tsx   # 问题卡片（统一渲染入口）
    │   │   ├── question-select.tsx # 单选题组件
    │   │   ├── question-multi-select.tsx # 多选题组件
    │   │   ├── question-boolean.tsx # 布尔题组件
    │   │   ├── question-text.tsx   # 文本题组件
    │   │   ├── question-code.tsx   # 代码题组件
    │   │   ├── question-image.tsx # 图片题组件
    │   │   ├── question-file.tsx   # 文件题组件
    │   │   ├── question-slider.tsx # 滑块题组件
    │   │   ├── question-rate.tsx   # 评分题组件
    │   │   ├── question-rank.tsx   # 排序题组件
    │   │   ├── question-options.tsx # 选项题通用组件
    │   │   ├── question-plan.tsx   # 计划题组件
    │   │   ├── question-review.tsx # 审查题组件
    │   │   ├── question-diff.tsx   # Diff 对比题组件
    │   │   ├── detail-panel.tsx    # 详情面板
    │   │   ├── history-card.tsx    # 历史卡片
    │   │   ├── debug-dialog.tsx    # 调试对话框
    │   │   ├── motion-demo-panel.tsx # 动画演示面板
    │   │   └── mock-data.ts        # 模拟数据（开发用）
    │   └── ui/                 # shadcn/ui 组件（通过 CLI 添加）
    ├── hooks/                  # React Hooks
    │   ├── useAuth.ts          # 认证 Hook（登录/登出/刷新/初始化/自动续期）
    │   ├── useApikey.ts        # API Key 数据 Hook（CRUD + 分页）
    │   ├── useProject.ts       # 项目数据 Hook（CRUD + 分页）
    │   ├── useQaAdmin.ts       # Q&A 管理 Hook（会话列表/详情/删除/配置）
    │   ├── useQaSession.ts     # Q&A 会话 Hook（问题状态管理 + 回答提交）
    │   ├── useQaWebSocket.ts   # Q&A WebSocket Hook（连接管理 + 消息回调）
    │   └── use-mobile.ts       # 移动端检测 Hook
    └── lib/
        ├── utils.ts            # cn() 工具（clsx + tailwind-merge）
        ├── motion.ts           # motion 动画变体和缓动函数
        ├── apis/               # API 客户端层
        │   ├── client.ts       # axios 实例（拦截器：自动附加 Token、401 清理跳转）
        │   ├── auth.ts         # 认证 API
        │   ├── apikey.ts       # API Key API
        │   ├── project.ts      # 项目 API
        │   └── qa-admin.ts     # Q&A 管理 API（会话/问题/配置）
        └── models/             # TypeScript 类型定义
            ├── request/        # 请求 DTO
            │   ├── auth.ts
            │   ├── apikey.ts
            │   ├── project.ts
            │   └── qa-admin.ts # Q&A 请求参数（SessionListParams/UpdateQaConfigRequest）
            └── response/       # 响应 DTO
                ├── common.ts   # BaseResponse 通用响应
                ├── page.ts     # 分页响应
                ├── auth.ts
                ├── apikey.ts
                ├── project.ts
                └── qa-admin.ts # Q&A 响应类型（Session/Question/Config）
```

## 导航指南

| 任务 | 位置 | 说明 |
|---|---|---|
| 新增页面 | `src/routes/` | 文件路径即路由路径；布局路由以 `_` 前缀 |
| 新增控制台子页面 | `src/routes/console/` | 在 `console.tsx` 布局下添加，自动继承 Sidebar + Breadcrumb |
| 新增 Interact 子页面 | `src/routes/interact/` | 在 `interact.tsx` 布局下添加 |
| 新增布局路由 | `src/routes/<name>.tsx` | 含 `Outlet` 的布局组件 |
| 新增通用组件 | `src/components/` | Navbar/Footer/Sidebar 级别的全局组件 |
| 新增业务组件 | `src/components/<domain>/` | 按业务域组织（如 apikey/、project/、qa/、interact/） |
| 新增题型组件 | `src/components/interact/question-*.tsx` | 遵循 `question-<type>.tsx` 命名 |
| 新增 shadcn/ui 组件 | `src/components/ui/` | 通过 `pnpm dlx shadcn@latest add <name>` |
| 新增 API 接口 | `src/lib/apis/` | 使用 apiClient 封装，返回类型化响应 |
| 新增数据 Hook | `src/hooks/` | 基于 TanStack Query 的 useMutation/useQuery |
| 新增类型定义 | `src/lib/models/` | 按 request/response 子目录组织 |
| 修改全局主题色 | `src/styles.css` | 仅修改 CSS 变量和 `@theme inline` |
| 修改路由配置 | `src/router.tsx` | 预加载策略、滚动恢复等 |
| 修改动画配置 | `src/lib/motion.ts` | 缓动函数和全局动画变体 |
| 工具函数 | `src/lib/` | 通用工具（如 `cn()`） |

## 约定

- **包管理器为 pnpm**：禁止使用 npm 或 yarn，确保 lock 文件一致性。
- **代码风格**：Prettier（`semi: false`、`singleQuote: true`、`trailingComma: "all"`）+ ESLint（`@tanstack/eslint-config`）。
- **路径别名**：`#/*` 映射到 `./src/*`；组件内使用 `#/components/xxx` 导入。
- **路由模式**：TanStack Start file-router；文件名即路由路径，`_` 前缀为布局路由。
- **认证守卫**：`__root.tsx` 通过 `beforeLoad` 检查初始化状态并自动重定向；`console.tsx` 通过 Cookie 检查 `access_token` 守卫。
- **Token 管理**：使用 `js-cookie` 存储 AT/RT/expires_at，`useAuth` Hook 每 30 秒检查并在 AT 过期前 5 分钟自动续期。
- **API 客户端**：`apiClient`（axios）自动附加 Bearer Token，响应拦截器处理业务错误码和 401 跳转。
- **数据获取**：使用 TanStack Query（useQuery/useMutation），缓存和请求状态由 QueryClient 管理。
- **shadcn/ui 管理**：组件通过 `pnpm dlx shadcn@latest add <component>` 添加，禁止手动创建 `ui/` 下的文件。
- **CSS 架构**：`styles.css` 仅负责 CSS 变量定义、`@theme inline` 映射、body 基础样式、`@layer base` 全局约束、`prefers-reduced-motion` 降级。组件级样式由组件自身通过 Tailwind 类管理。
- **动画库**：使用 `motion`（Framer Motion 的轻量版）实现动画；缓动函数和变体统一定义在 `lib/motion.ts`。
- **主题配色**：微明色盘（烛光暖褐系），亮/暗模式通过 `:root` / `.dark` CSS 变量切换；shadcn/ui 变量已兼容主题。
- **Toast 通知**：使用 `sonner`（shadcn/ui 集成），在 `console.tsx` 布局中挂载 `<Toaster />`。
- **前端嵌入**：构建产物输出到 `web/dist`，通过 `go:embed` 嵌入 Go 二进制实现单文件部署。
- **Q&A 实时通信**：Interact 页面通过 WebSocket 与后端通信，`useQaWebSocket` 管理连接状态和消息分发。
- **Q&A 管理端**：Console Q&A 页面通过 REST API 管理会话，使用 `useQaAdmin` Hook。
- **题型组件**：Interact 页面每种题型对应独立的 `question-<type>.tsx` 组件，通过 `question-card.tsx` 统一分发渲染。

## 反模式

- 禁止在 `styles.css` 中编写组件级或页面级样式。
- 禁止手动编辑 `routeTree.gen.ts`，它由 TanStack Router 插件自动生成。
- 禁止手动创建 `src/components/ui/` 下的 shadcn/ui 组件文件。
- 禁止使用 npm 或 yarn 安装依赖。
- 禁止在 `styles.css` 中引入第三方 CSS 库的完整样式。
- 禁止直接操作 `localStorage` 存储认证凭据；统一使用 `js-cookie`。
- 禁止在组件中直接调用 axios；必须通过 `lib/apis/` 层封装。

## 调试路径

1. 页面 404 → 检查 `src/routes/` 文件路径是否正确匹配路由。
2. 认证循环重定向 → 检查 `__root.tsx` 的 `beforeLoad` 逻辑和 `getStatus` API 返回。
3. 控制台空白/未授权 → 检查 Cookie 中 `access_token` 是否存在，`console.tsx` 的 `beforeLoad` 守卫。
4. 样式异常 → 检查 `styles.css` 的 CSS 变量是否被覆盖；确认组件使用 Tailwind 类而非自定义 CSS。
5. shadcn/ui 组件不显示 → 确认通过 CLI 正确安装，检查 `components.json` 别名配置。
6. 路由跳转失败 → 检查 `router.tsx` 配置和 `routeTree.gen.ts` 是否为最新。
7. 动画不播放 → 确认 `motion` 导入是否正确（`motion/react`）；检查 `prefers-reduced-motion` 设置。
8. API 请求 401 → 检查 `lib/apis/client.ts` 拦截器是否正确附加 Token，Cookie 是否过期。
9. 数据表格不刷新 → 检查 `useQuery` 的 `queryKey` 和 `staleTime` 配置。
10. Q&A WebSocket 断连 → 检查 `useQaWebSocket.ts` 连接状态和后端 `route_ws.go` 端点。
11. Interact 题型渲染异常 → 检查 `components/interact/question-card.tsx` 的题型分发逻辑。
