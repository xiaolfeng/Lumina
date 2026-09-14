# Preview 与 Pages 区分架构与界面设计

> 状态：draft · 承接定稿 [ADR-0007](../adr/0007-preview-pages-separation.md) · 关联 [ADR-0001](../adr/0001-architecture-runtime-boundaries.md) · 关联 [ADR-0002](../adr/0002-project-identity-resolution.md)

| 项 | 值 |
| --- | --- |
| 作者 | 筱锋 (xiao_lfeng) / Lumina Architecture Team |
| 日期 | 2026-09-14 |
| 状态 | Draft |
| 范围词 | 预览工作台 (`preview`) / 即时页面 (`pages`) / 快照晋升 (`promotion`) / 路径式寻址 / 版本闭环 |

---

## 一、概述 (Overview)

在 Lumina 原有架构中，`Preview` 模块专注于 Agent 与开发者之间轻量、即时、交互式的原型走查。但在实际团队协作中，大量经过多轮打磨的原型（如设计规范、交互组件库、技术白皮书演练、微型数据看板）已具备长久留存与对外分享的价值。

若直接将原有临时 Preview 会话作为持久化托管，会导致**生命周期耦合**（7 天 TTL 定期清理会误删正式成果）、**路径解析断裂**（Query 参数 `?session=...&file=...` 破坏了 Web 相对路径标准）、以及**交付体验割裂**（调试工具条与内部代码树强制暴露给最终读者）。

本设计确立了同一业务域下的双层演化模型：
1. **Preview（调试工作台 / Workbench）**：会话级、易失性草稿，保留完整侧边栏（Sidebar）、多端型视口切换与实时调试能力；
2. **Pages（持久即时页面 / Showcase）**：项目级不可变发布态，采用语义化路径寻址；界面采用沉浸式画布，并通过专属的**右侧抽屉**实现 `.html` / `.md` 页面一键直切，以及底层 `.css` / `.js` 资源折叠查阅的「高级」面板；
3. **晋升机制（Promotion）**：一键从 Preview 生成不可变快照并绑定 Project 发布为 Pages；
4. **全路径式寻址（Path-based Routing）**：全面替换 Query 参数，原生支持相对资源引用与子页面无缝跳转；
5. **版本闭环（Fork-Promotion-Active Pointer）**：通过溯源元数据透传、智能递增晋升和版本生效指针，实现安全的迭代发布与秒级无损回滚。

---

## 二、背景与痛点 (Background & Motivation)

### 2.1 痛点 1：Query 寻址破坏相对路径标准
原有访问形式为 `/preview?session=<hash>&file=index.html`。当 HTML 文件内部存在相对链接（如 `<a href="about.html">`）或静态资源依赖（如 `<link href="style.css">`、`<img src="./assets/banner.svg">`）时，浏览器的相对路径解析基准会直接挂在 `/preview` 之下，而非当前会话目录，导致点击链接跳转时丢弃 query 参数或产生跨级 404，极难在不侵入用户 HTML 代码的前提下正常运行多页面原型。

### 2.2 痛点 2：草稿生命周期与正式成果的冲突
`PreviewSession` 具有 TTL 回收机制（默认最大存活 7 天），且支持开发者手动批量清理。当原型达到可交付状态后，若无持久化载体，过期清理机制将直接摧毁生产依赖，造成「不敢清理预览」与「持久页面被误删」的双向矛盾。

### 2.3 痛点 3：交付视觉心智错乱
Preview 界面是为调试设计的：左侧强制展示文件列表树，顶部强制展示会话 ID、刷新频率与草稿状态。当把该链接分享给产品经理、设计师或外部客户时，充斥着开发噪音，无法作为专业的独立作品展示。

### 2.4 痛点 4：版本迭代与更新缺乏闭环
当页面发布后，若后续需要修改升级，必须从线上内容重新派生（Fork）出可调试的草稿。若缺乏版本溯源与生效指针机制，修改后的草稿保存时无法确定该覆盖哪个版本、如何生成新版本，以及如何确保线上受众安全、平滑地看到更新后的生效版本。

---

## 三、目标与非目标 (Goals & Non-Goals)

### 3.1 核心目标 (Goals)
- **路径式寻址标准化**：
  - Preview 采用 `/preview/:session_hash/*filepath`；
  - Pages 采用 `/pages/:project_name/:slug/*filepath`；
  - 自动识别缺省路径并 302 重定向至入口 HTML。
- **不可变快照晋升 (Promotion)**：
  - 支持从 Preview 会话全量深拷贝文件生成不可变版本快照（`v1.0.0`），存入 Pages 持久表；
  - Preview 会话的删除或清理对已发布 Pages 绝对零影响。
- **解耦的 UI 架构**：
  - **Preview Workbench**：保留左侧文件树 Sidebar、顶部设备视口切换器（Desktop/Tablet/Mobile）、源码检查与「晋升为 Pages」入口；
  - **Pages Showcase**：默认全屏沉浸渲染；右上角提供可折叠的**微明悬浮胶囊（Showcase Pill）**；
  - **Pages 专属交互右侧抽屉**：
    - **直接渲染页面直切区**：突出展示 `.html` 与 `.md` 文件清单，点击直接在主画布上切换渲染；
    - **高级资源折叠区 (Advanced)**：默认折叠收纳 `.css`、`.js`、`.json` 等底层辅助资源，点击可调出代码检查器。
- **严密自洽的版本迭代闭环**：
  - Fork 操作完整注入 `source_page_id` 与 `source_version_id` 溯源标记；
  - 晋升弹窗具备双态识别：未绑定源走「创建新 Page」，已绑定源走「发布新版本（递增版本号 + 更新说明 + 立即生效开关）」；
  - 采用单一指针 `Page.latest_version_id` 管理线上渲染，支持秒级切换生效版本与无损回滚。

### 3.2 明确非目标 (Non-Goals)
- **非目标 1**：v1 不引入外部自定义独立域名（CNAME 解析与证书管理），统一依托 Lumina 主域名子路径分发。
- **非目标 2**：不引入服务端动态脚本或模板执行引擎（如 PHP/JSP/SSR），维持纯静态沙盒（`iframe sandbox="allow-scripts"`）。
- **非目标 3**：不将草稿的 WebSocket 实时同步机制直接套用至 Pages；Pages 必须通过显式晋升保障线上高可用。

---

## 四、核心数据模型设计 (Data Architecture)

在实体层，新增 3 个持久化实体，并在已有 `PreviewSession` 上补充溯源外键：

```text
Project (已有, Gene=32)
   │
   ├── PreviewSession (已有, Gene=45, 易失性草稿) ──[Promotion 晋升]──┐
   │      ├── SourcePageID (可选, 溯源目标 Page)                   │
   │      ├── SourceVersionID (可选, 溯源基准版本)                 │
   │      └── PreviewFile (已有, Gene=46)                          │
   │                                                               │
   └── Page (新增, Gene=49, 语义化持久实体) ◀──────────────────────┘
          │
          ├── LatestVersionID (当前线上生效快照指针)
          │
          └── PageVersion (新增, Gene=50, 不可变快照版本序列)
                 │
                 └── PageFile (新增, Gene=51, 快照文件内容)
```

### 4.1 实体字段详细定义

1. **`PreviewSession` 实体扩展（Gene = 45）**：
   - 追加 `SourcePageID` (`*SnowflakeID`): 若从 Pages Fork 而来，记录关联的 Page ID；
   - 追加 `SourceVersionID` (`*SnowflakeID`): 记录 Fork 时所基于的快照版本 ID。

2. **`Page` 实体（Gene = 49）**：
   - `ID` (`SnowflakeID`): 主键；
   - `ProjectID` (`SnowflakeID`): 所属项目 ID；
   - `Slug` (`varchar(64)`): 项目内唯一的访问标识（正则：`^[a-z0-9-]+$`）；
   - `Title` (`varchar(255)`): 页面显示标题；
   - `Description` (`text`): 页面描述；
   - `Status` (`varchar(16)`): 状态（`published` / `archived`）；
   - `AccessMode` (`varchar(16)`): 访问权限策略（`public` / `password`）；
   - `PasswordHash` (`varchar(128)`): 访问密码哈希（可为空，密码保护模式下必填）；
   - `LatestVersionID` (`SnowflakeID`): **当前线上生效的版本指针**（对外服务单一事实来源）；
   - 联合唯一索引：`uk_project_slug (project_id, slug)`。

3. **`PageVersion` 实体（Gene = 50）**：
   - `ID` (`SnowflakeID`): 版本主键；
   - `PageID` (`SnowflakeID`): 所属 Page ID；
   - `Version` (`varchar(32)`): 语义化版本号（如 `v1.0.0`、`v1.1.0`）；
   - `Changelog` (`text`): 该版本的更新日志；
   - `SourceSessionID` (`*SnowflakeID`): 来源 Preview 会话 ID（用于审计溯源）；
   - `BaseVersionID` (`*SnowflakeID`): 派生基准版本 ID（用于冲突检测）；
   - `EntryFilename` (`varchar(255)`): 默认入口文件名（默认为 `index.html`）；
   - `FileCount` (`int`): 快照内文件总数；
   - `TotalSize` (`int64`): 快照总字节数；
   - `CreatedBy` (`varchar(64)`): 发布操作者。

4. **`PageFile` 实体（Gene = 51）**：
   - `ID` (`SnowflakeID`): 文件主键；
   - `VersionID` (`SnowflakeID`): 所属快照版本 ID；
   - `Filename` (`varchar(255)`): 扁平单层文件名（严格禁止包含 `/`、`\`、`..`）；
   - `MimeType` (`varchar(128)`): 文件 MIME 类型；
   - `Size` (`int`): 文件大小（上限 256 KiB）；
   - `Content` (`text`): UTF-8 文件正文内容。

---

## 五、寻址与路由规范 (Routing & URL Resolution)

全面推行全路径式匹配（Path-based Routing），前后端协同路由映射如下：

| 功能场景 | 外部访问 URL | 内部路由处理机制 | 行为表现 |
|---|---|---|---|
| **Preview 根路径** | `/preview/:session_hash` 或 `/preview/:session_hash/` | 强制 `middleware.Auth` 登录鉴权；检查会话 `entry_file`，302 跳转至 `/preview/:session_hash/index.html` | 保持浏览器地址栏直观明确；未登录重定向至登录页 |
| **Preview 文件渲染** | `/preview/:session_hash/:filename` | 强制 `middleware.Auth` 登录鉴权；从 `PreviewFile` 读取对应文件内容返回 | 沙盒 iframe 内部同层链接相对跳转原生生效 |
| **Pages 根路径** | `/pages/:project_name/:slug` 或 `/pages/:project_name/:slug/` | 经过 `middleware.PagesAuth` 密码门校验；查询当前 `LatestVersion` 的 `EntryFilename`，302 跳转 | 保持规范化永久入口；未解锁密码门时呈现密码验证卡片 |
| **Pages 文件渲染** | `/pages/:project_name/:slug/:filename` | 经过 `middleware.PagesAuth` 密码门校验；从 `PageFile` 读取该版本对应文件，按 MIME 输出 | 原生支持子页面切换与资源加载 |
| **Pages 指定版本访问** | `/pages/:project_name/:slug/:filename?v=1.0.0` | 经过 `middleware.PagesAuth` 密码门校验；按指定历史版本读取 `PageFile` 渲染 | 供团队历史对照与回溯查阅 |

### 5.1 页面内相对链接原生支持与地址栏双向同步 (Relative Links Navigation)
在旧版 Query 参数模式下，沙盒内页面的相对链接（如 `<a href="components.html">`）在点击时无法正常被外部地址栏感知，导致用户刷新后依然回到主入口，无法书签化或分发特定子页面。

在全路径寻址下，确立以下相对链接交互契约：
1. **沙盒与外层地址栏双向无刷新联动**：
   - 读者在页面内点击任意相对跳转（`.html`、`.md`、`.htm`），前端路由监听层捕获该导航事件；
   - 主画布平滑渲染目标页面，同时无刷新将外部浏览器地址栏与工具条的路径更新为对应目标文件（如 `/preview/7f8a92b104c8/components.html` 或 `/pages/lumina/design-system/components.html`）；
2. **多端状态同步**：
   - 在 Preview 工作台模式下，页面内相对跳转同步更新左侧 Sidebar 的激活高亮状态；
   - 在 Pages 展示态下，页面内相对跳转同步更新右侧抽屉的可渲染页面高亮。

---

## 六、版本迭代与 Fork-Promotion 闭环机制

针对「线上已生效版本 → Fork 修改 → 保存晋升 → 设为新版本生效」的完整生命周期，设计如下四步闭环契约：

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. 线上 Pages 展示态 (Showcase)                                             │
│    - 当前线上生效版本: v1.0.0 (Page.LatestVersionID -> ver_100)             │
│    - 点击右侧微明胶囊中的「⎇ Fork 到新预览」                                   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ POST /api/v1/pages/:id/fork
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. 派生 Preview 调试工作台 (Workbench)                                      │
│    - 注入溯源: source_page_id = page_123, source_version_id = ver_100       │
│    - 顶部状态栏显式提醒: "🔄 正在迭代: lumina/design-system (基于 v1.0.0)"   │
│    - 开发者/Agent 进行多轮代码修改、WebSocket 热刷新走查                    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ 验证满意，点击「🚀 晋升为 Pages」
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. 智能晋升双态对话框 (Context-Aware Promotion Modal)                        │
│    - 系统自动检测到关联来源 Page: lumina / design-system (锁定不可篡改)     │
│    - 自动推荐递增版本号: v1.1.0 (允许微调)                                   │
│    - 填写更新说明 (Changelog): "优化色彩对比度与新增 Buttons 规范"          │
│    - 默认勾选: [✓] 立即设为当前线上生效版本 (Make Active Immediately)        │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ POST /api/v1/preview/sessions/:id/promote
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 4. 原子生效与不可变快照落库                                                  │
│    - 事务中生成新版本 ver_110 (PageVersion) 与全量 PageFiles                │
│    - 若勾选立即生效: 原子更新 Page.LatestVersionID = ver_110                 │
│    - 线上路由 /pages/lumina/design-system/* 瞬间平滑无缝切换至 v1.1.0       │
│    - 历史快照 ver_100 永久归档，支持随时在控制台一键无损回滚                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.1 并发冲突防护 (Optimistic Concurrency Control)
当开发者 B 在基于 `v1.0.0` 进行修改时，如果开发者 A 已经先一步将 `v1.1.0` 发布并设为线上生效版本：
- 开发者 B 提交晋升时，后端比对 `Page.LatestVersionID != PreviewSession.SourceVersionID`；
- 系统弹出冲突提示框：
  > 「⚠️ 线上已有更新版本发布：当前线上生效版本已推进至 `v1.1.0`（更新人：Alice，更新时间：10 分钟前）。您当前基于 `v1.0.0` 提交，是否确认继续发布为 `v1.2.0` 并覆盖线上生效指针？」
- 避免后发者的修改在团队无感知的情况下悄悄覆盖线上状态。

### 6.2 秒级无损回滚机制 (Instant Zero-Copy Rollback)
由于所有版本的 `PageFile` 均为静态不可变快照，如果新发布的 `v1.1.0` 在生产环境中被发现存在重大问题：
- 管理员只需在控制台或 Pages 胶囊面板中点击历史版本 `v1.0.0` 旁的 **「回滚至此版本」**；
- 后端仅需执行单条 SQL：`UPDATE pages SET latest_version_id = 'ver_100' WHERE id = 'page_123';`；
- 零文件深拷贝，耗时小于 5ms，线上公共访问即刻平滑回滚至 `v1.0.0`。

---

## 七、UI/UX 详细交互规格 (Interface & Interaction)

### 7.1 Preview 调试工作台 (Workbench)
- **左侧文件树 Sidebar**：
  - 完整保留垂直文件树，直观查看所有代码资产；
  - 标识 `[入口]` 徽章与文件体积；
  - 底部显示路径式路由格式、WebSocket 连接状态与 TTL 倒计时；
- **顶部调试栏 (Toolbar)**：
  - 左侧：若为 Fork 派生会话，显示带有双向箭头的 **「🔄 正在迭代：lumina/design-system (基于 v1.0.0)」** 提示标签；
  - 中间：设备视口切换按钮组（`桌面 100%`、`平板 768px`、`移动端 375px`）；
  - 右侧：视口渲染 / 源码检查模式切换，以及显式强调的 **「🚀 晋升为 Pages」** 操作按钮；
- **画布视口容器与双态切换 (Canvas Viewport)**：
  - **视口渲染态**：
    - **桌面模式 (Desktop 100%)**：无外围边框（`border: none`）、无外围投影（`box-shadow: none`）、零边距全平铺，与页面背景天然融为一体，彻底消除多余的外框束缚与视觉损耗；
    - **平板与移动端模式 (Tablet 768px / Mobile 375px)**：
      - 画布自动激活底衬沉浸模式（`background: #e9e3d8`），顶部预留 36px 优雅的呼吸留白（避免与上方工具栏发生视觉粘连）；
      - 设备容器呈现精致的 1px 细线框与立体悬浮阴影（`box-shadow: 0 12px 36px rgba(0,0,0,0.14)`），犹如实体设备卡片悬浮于画布中央；
      - **内部 y 轴全高度自适应最大化上下滚动**：设备外框高度自动撑满画布可用垂直空间的最高上限（`flex: 1; height: calc(100% - 38px); min-height: 0;`），消除多余的纵向死黑空隙；内部渲染表面独立接管垂直滚动（`overflow-y: auto; overflow-x: hidden`），搭配细窄滚动条，精准还原移动端与平板的真实单机滚动体验；
      - **自由拖拽缩放与底部常驻 Tip 提示条**：容器右侧边缘配置拖拽拉伸把手（Resize Handle），允许开发者按住把手左右自由拖拽调整宽度（320px 至 1200px）；设备正下方独立设置常驻 Tip 提示条（如 `↔ 宽度: 375px (手机) | 按住右侧把手自由拖拽 · y 轴全高度自适应滚动`），独立于滚动容器之外，绝不被任何内部 overflow 截断；
  - **源码检查态 (Source Code Inspector)**：
    - 开发者点击「源码检查」时，**界面默认自动缩回桌面全宽平铺，并动态隐藏顶部「桌面 / 平板 / 手机」端型切换器**；
    - 源码编辑器独占工作台整个宽度与高度，横向排版与代码行宽得到完整舒展；
    - 切回「视口渲染」时，系统自动无感恢复端型切换控制与原视口尺寸。

### 7.2 Pages 持久展示态 (Showcase)
- **100% 沉浸全屏画布**：
  - 完全隐藏调试工具栏与常驻文件树，保证设计成果的原汁原味呈现；
- **右上角常驻可拖拽微明胶囊 (Draggable Showcase Pill)**：
  - 极简常驻：`⋮⋮ ✦ lumina / design-system v1.0.0 ▾`；
  - **自由拖拽避让**：胶囊自带抓手手柄，支持在屏幕任意区域自由按住拖拽移动，彻底避免遮挡特定页面的关键内容或右上方控制按钮；
  - **防误触智能识别**：内置 4px 移动阈值检测，位移 > 4px 仅执行位置重设，位移 <= 4px 单击则平滑展开右侧交互抽屉；视口边缘防逃逸保护；
- **右侧展开交互抽屉 (Slide-over Drawer)**：
  点击悬浮胶囊滑出抽屉，由上至下清晰分层：
  1. **元信息区**：项目归属、不可变快照状态、当前生效版本（含版本历史切换入口）；
  2. **可直接渲染页面直切区（Directly Renderable Pages）**：
     - 自动筛选当前快照中的 `.html` 和 `.md` 文件；
     - 以卡片形式清晰罗列页面标题与文件名；
     - **核心交互**：点击任一页面卡片，主画布**直接无刷新切换为该页面的渲染视图**（HTML 走沙盒 iframe，Markdown 走微明排版引擎），地址栏路径同步变更为对应子路径；
  3. **高级资源折叠区（高级 · Advanced Assets）**：
     - 以手风琴折叠框呈现，默认折叠收纳 `.css`、`.js`、`.json`、`.svg` 等底层辅助资源；
     - 展开后点击任一文件，弹出全屏代码检查面板查阅源码，满足极客与开发者走查需求，同时避免干扰普通读者；
  4. **底部快捷动作**：
     - 「复制当前路径」：复制对应子页面的完整公共 URL；
     - 「⎇ Fork 到新预览」：一键从当前版本深拷贝创建全新草稿会话并进入 Workbench。

---

## 八、访问控制与密码门安全机制 (Access Control & Password Gate)

### 8.1 Preview 内部草稿强制登录鉴权
- **研发资产定位**：Preview 是面向 Agent 与平台开发者的内部调试沙盒，包含完整目录树、未发布的代码与内部会话 Hash，属于平台核心受保护资源；
- **鉴权约束**：必须挂载 `middleware.Auth`（控制台 Bearer Token / 会话认证）；未登录用户访问任何 `/preview/:session_hash/*` 路由时，一律拦截并返回 401 或重定向至控制台登录页 `/_public/login`，严禁研发草稿在公开互联网裸露。

### 8.2 Pages 持久交付态可选密码门机制 (Optional Password Gate)
- **交付场景诉求**：Pages 作为正式交付物，需兼顾「完全公开对外分发」与「团队/客户私密阶段性交付」两类典型场景；
- **配置落点与职责隔离（核心架构原则）**：
  - **严禁在「晋升为 Pages」的轻量弹窗中配置密码**：晋升动作是开发者或 Agent 针对当前验证通过的代码生成快照的纯粹工程行为，必须保持轻量、极简且无心智负担；
  - **密码保护与访问策略统一在管理员端口界面（`/console/pages`）进行配置与治理**：
    - 控制台页面详情/安全设置抽屉中，管理员可对已发布的 Page 进行全生命周期的权限治理；
    - 管理端界面提供两类访问策略切换：
      1. `public`（公开访问）：默认模式，免登录、免密码，任何持有 URL 的访客均可直接浏览；
      2. `password`（密码保护）：设置或修改访问密码，系统通过加盐哈希（`bcrypt`）落库至 `Page.PasswordHash`；
    - 支持密码重置、清除密码恢复公开、失效历史会话等完整的运维管控能力；
- **密码门拦截与无感会话流程**：
  1. 访客首次访问受密码保护的 `/pages/:project/:slug/*` 时，`middleware.PagesAuth` 拦截请求并渲染微明标准「密码门」界面（展示输入框与解锁按钮）；
  2. 访客输入密码提交至 `POST /api/v1/pages/by-project/:project/:slug/unlock`；
  3. 后端比对密码哈希通过后，签发有时效签名的安全 HttpOnly Cookie（`lum_pages_auth`，默认 24 小时有效）；
  4. 访客浏览器在有效期内可免密畅通访问该快照下的全量 HTML、Markdown 以及相对引用的 CSS、JS、图片资产，无需重复输入密码；
  5. 整体架构复用成熟的 `WikiAuth` 密码门设计，保持系统底层安全模型统一。

---

## 九、REST API 矩阵 (API Specification)

### 9.1 Preview 模块接口
- `GET /api/v1/preview/sessions/:hash/*filepath`：按路径读取会话内指定文件内容（需登录）；
- `POST /api/v1/preview/sessions/:id/promote`：
  - 请求体：
    ```json
    {
      "slug": "design-system",
      "title": "微明统一设计系统与规范",
      "description": "设计系统说明",
      "version": "v1.1.0",
      "changelog": "优化色彩语义与按钮组件",
      "set_as_active": true
    }
    ```

### 9.2 Pages 模块接口
- `GET /api/v1/pages/:project_name/:slug/*filepath`：按路径读取当前线上生效版本的文件（公开或密码校验）；
- `POST /api/v1/pages/by-project/:project_name/:slug/unlock`：公开接口，提交密码解锁密码门，返回 Session Cookie；
- `PUT /api/v1/pages/:id/access-policy`：管理员接口（需控制台登录），修改页面访问策略与密码保护：
  ```json
  {
    "access_mode": "password",
    "password": "your_secure_password"
  }
  ```
- `GET /api/v1/pages/:id/versions`：获取该页面的所有历史快照版本列表；
- `POST /api/v1/pages/:id/versions/:version_id/switch-active`：切换线上当前生效版本指针；
- `POST /api/v1/pages/:id/fork`：
  - 请求体：`{ "version_id": "<可选，默认当前生效版本>" }`；
  - 响应：返回新创建的 `PreviewSession` 详情与跳转路径。
