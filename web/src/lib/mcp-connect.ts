/**
 * MCP 接入配置的单一数据源。
 * 工具名必须与 internal/mcp/{qa,project,pin,repowiki,preview,pages,workspace}_tools.go 保持一致。
 */

export const MCP_PATH = '/api/v1/mcp'
export const MCP_KEY_PLACEHOLDER = '<api-key>'
export const MCP_SERVER_NAME = 'lumina'

export interface McpConnectContext {
  mcpUrl: string
  apiKey?: string | null
}

export type McpSnippetLanguage = 'json' | 'toml' | 'bash' | 'text'

export interface McpClientDef {
  id: string
  name: string
  fileHint: string
  language: McpSnippetLanguage
  note: string
  build: (ctx: McpConnectContext) => string
}

export interface McpToolDef {
  name: string
  summary: string
}

export interface McpToolModule {
  id: string
  name: string
  summary: string
  note?: string
  tools: McpToolDef[]
}

export interface McpWorkflowStep {
  step: number
  title: string
  body: string
  tools: string[]
}

export function trimSlash(origin: string): string {
  return origin.trim().replace(/\/+$/, '')
}

export function resolveMcpOrigin(options: {
  override?: string | null
  siteDomain?: string | null
  windowOrigin?: string | null
}): string {
  const override = trimSlash(options.override ?? '')
  if (override) return override
  const siteDomain = trimSlash(options.siteDomain ?? '')
  if (siteDomain) return siteDomain
  return trimSlash(options.windowOrigin ?? '')
}

export function buildMcpUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return MCP_PATH
  return `${base}${MCP_PATH}`
}

export function normalizeApiKey(apiKey?: string | null): string {
  const trimmed = apiKey?.trim()
  return trimmed ? trimmed : MCP_KEY_PLACEHOLDER
}

export function buildAuthorizationValue(apiKey?: string | null): string {
  return `Bearer ${normalizeApiKey(apiKey)}`
}

export function buildAuthorizationHeader(apiKey?: string | null): string {
  return `Authorization: ${buildAuthorizationValue(apiKey)}`
}

function jsonBlock(value: unknown): string {
  return JSON.stringify(value, null, 2)
}

function tomlQuote(value: string): string {
  return `"${value.replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`
}

function isPlainHttp(url: string): boolean {
  return url.startsWith('http://')
}

function cursorSnippet(ctx: McpConnectContext): string {
  return jsonBlock({
    mcpServers: {
      [MCP_SERVER_NAME]: {
        url: ctx.mcpUrl,
        headers: {
          Authorization: buildAuthorizationValue(ctx.apiKey),
        },
      },
    },
  })
}

function windsurfSnippet(ctx: McpConnectContext): string {
  return jsonBlock({
    mcpServers: {
      [MCP_SERVER_NAME]: {
        serverUrl: ctx.mcpUrl,
        headers: {
          Authorization: buildAuthorizationValue(ctx.apiKey),
        },
      },
    },
  })
}

function vscodeSnippet(ctx: McpConnectContext): string {
  return jsonBlock({
    servers: {
      [MCP_SERVER_NAME]: {
        type: 'http',
        url: ctx.mcpUrl,
        headers: {
          Authorization: buildAuthorizationValue(ctx.apiKey),
        },
      },
    },
  })
}

function clineSnippet(ctx: McpConnectContext): string {
  return jsonBlock({
    mcpServers: {
      [MCP_SERVER_NAME]: {
        disabled: false,
        type: 'streamableHttp',
        url: ctx.mcpUrl,
        headers: {
          Authorization: buildAuthorizationValue(ctx.apiKey),
        },
      },
    },
  })
}

function claudeDesktopSnippet(ctx: McpConnectContext): string {
  const args = ['-y', 'mcp-remote', ctx.mcpUrl]
  if (isPlainHttp(ctx.mcpUrl)) args.push('--allow-http')
  args.push('--header', 'Authorization:${AUTH_HEADER}')
  return jsonBlock({
    mcpServers: {
      [MCP_SERVER_NAME]: {
        command: 'npx',
        args,
        env: {
          AUTH_HEADER: buildAuthorizationValue(ctx.apiKey),
        },
      },
    },
  })
}

/**
 * OAuth 直连（方案 A）：不带 Authorization 头，由客户端在首次连接时
 * 发起 OAuth 2.1 登录授权，Lumina 签发并自动续期访问令牌。
 */

export function buildClaudeOAuthAdd(mcpUrl: string): string {
  return `claude mcp add --transport http ${MCP_SERVER_NAME} ${mcpUrl}`
}

export function buildCodexOAuthAdd(mcpUrl: string): string {
  return `codex mcp add ${MCP_SERVER_NAME} --url ${mcpUrl}`
}

export function buildGenericOAuthJSON(mcpUrl: string): string {
  return JSON.stringify(
    { mcpServers: { [MCP_SERVER_NAME]: { type: 'http', url: mcpUrl } } },
    null,
    2,
  )
}

function claudeCodeSnippet(ctx: McpConnectContext): string {
  return `claude mcp add --transport http ${MCP_SERVER_NAME} ${ctx.mcpUrl} --header "${buildAuthorizationHeader(ctx.apiKey)}"`
}

function grokCliSnippet(ctx: McpConnectContext): string {
  return `grok mcp add --transport http ${MCP_SERVER_NAME} ${ctx.mcpUrl} --header "${buildAuthorizationHeader(ctx.apiKey)}"`
}

function codexSnippet(ctx: McpConnectContext): string {
  return [
    `[mcp_servers.${MCP_SERVER_NAME}]`,
    `url = ${tomlQuote(ctx.mcpUrl)}`,
    `http_headers = { Authorization = ${tomlQuote(buildAuthorizationValue(ctx.apiKey))} }`,
  ].join('\n')
}

function grokSnippet(ctx: McpConnectContext): string {
  return [
    `[mcp_servers.${MCP_SERVER_NAME}]`,
    `url = ${tomlQuote(ctx.mcpUrl)}`,
    `headers = { Authorization = ${tomlQuote(buildAuthorizationValue(ctx.apiKey))} }`,
  ].join('\n')
}

function genericSnippet(ctx: McpConnectContext): string {
  return [
    '传输：Streamable HTTP',
    `端点：${ctx.mcpUrl}`,
    `鉴权：${buildAuthorizationHeader(ctx.apiKey)}`,
  ].join('\n')
}

export const MCP_CLIENTS: McpClientDef[] = [
  {
    id: 'cursor',
    name: 'Cursor',
    fileHint: '~/.cursor/mcp.json 或项目 .cursor/mcp.json',
    language: 'json',
    note: '保存后在 Cursor Settings → MCP 中刷新。可用 ${env:LUMINA_API_KEY} 代替明文密钥。',
    build: cursorSnippet,
  },
  {
    id: 'windsurf',
    name: 'Windsurf',
    fileHint: '~/.codeium/windsurf/mcp_config.json',
    language: 'json',
    note: 'Windsurf 使用 serverUrl，不要改成 Cursor 的 url 字段。',
    build: windsurfSnippet,
  },
  {
    id: 'vscode',
    name: 'VS Code',
    fileHint: '.vscode/mcp.json',
    language: 'json',
    note: 'VS Code 1.99+ / Copilot Agent。根键是 servers，且必须带 type: "http"。',
    build: vscodeSnippet,
  },
  {
    id: 'cline',
    name: 'Cline / Roo Code',
    fileHint: 'Cline：cline_mcp_settings.json',
    language: 'json',
    note: 'Roo Code 配置形态相同。type 使用 streamableHttp。',
    build: clineSnippet,
  },
  {
    id: 'claude-desktop',
    name: 'Claude Desktop',
    fileHint: '~/Library/Application Support/Claude/claude_desktop_config.json',
    language: 'json',
    note: 'Desktop 的 mcpServers 只跑本地 stdio，因此用 mcp-remote 桥接。Authorization 放在 env 里，避免 args 空格被截断。HTTP 明文地址会自动加上 --allow-http。',
    build: claudeDesktopSnippet,
  },
  {
    id: 'claude-code',
    name: 'Claude Code',
    fileHint: '终端命令',
    language: 'bash',
    note: '这是独立 MCP 接入方式。安装 Lumina 插件时无需执行；手动接入后可用 claude mcp list 确认。',
    build: claudeCodeSnippet,
  },
  {
    id: 'codex',
    name: 'Codex CLI',
    fileHint: '~/.codex/config.toml',
    language: 'toml',
    note: '根键是 mcp_servers（下划线）。也可用 bearer_token_env_var 代替明文 http_headers。',
    build: codexSnippet,
  },
  {
    id: 'grok',
    name: 'Grok',
    fileHint: '~/.grok/config.toml',
    language: 'toml',
    note: `也可执行：${grokCliSnippet({ mcpUrl: '<url>', apiKey: MCP_KEY_PLACEHOLDER })}`,
    build: grokSnippet,
  },
  {
    id: 'generic',
    name: '通用 HTTP',
    fileHint: '任意支持 Streamable HTTP 的客户端',
    language: 'text',
    note: '适用于支持 Streamable HTTP 的客户端。Q&A 的实时页面由 Lumina 自己通过 WebSocket 更新，无需另配实时服务。',
    build: genericSnippet,
  },
]

export function getMcpClient(id: string): McpClientDef | undefined {
  return MCP_CLIENTS.find((client) => client.id === id)
}

export function buildClientSnippet(id: string, ctx: McpConnectContext): string {
  const client = getMcpClient(id)
  if (!client) {
    throw new Error(`未知 MCP 客户端：${id}`)
  }
  return client.build(ctx)
}

export const MCP_TOOL_MODULES: McpToolModule[] = [
  {
    id: 'workspace',
    name: 'Workspace',
    summary: '先选定工作空间，再在该空间内解析项目。单用户用来拆开生活与工作。',
    tools: [
      { name: 'workspace_list', summary: '列出全部空间，默认空间 slug 为 default' },
      { name: 'workspace_get', summary: '按 ID 或 slug 查看空间' },
    ],
  },
  {
    id: 'project',
    name: 'Project',
    summary:
      '先认出当前代码库，让问答、预览和跨项目协作都落在正确的项目上下文中。',
    tools: [
      { name: 'project_get', summary: '按 ID、名称或路径找到已有项目' },
      { name: 'project_list', summary: '浏览项目，也能按当前目录筛选' },
      { name: 'project_create', summary: '当前代码库首次接入时创建项目' },
    ],
  },
  {
    id: 'qa',
    name: 'Q&A',
    summary:
      '遇到方案分支或需要补充材料时，Agent 可以把问题送到交互页，等你确认后再继续。',
    tools: [
      { name: 'qa_session_list', summary: '复用已有活跃会话' },
      { name: 'qa_session_create', summary: '创建会话，必须带 project_id' },
      { name: 'qa_session_get', summary: '查看会话详情' },
      { name: 'qa_session_archive', summary: '归档结束的会话' },
      { name: 'qa_what_question', summary: '查询当前待回答问题' },
      { name: 'qa_push_question', summary: '推送问题到交互页' },
      { name: 'qa_push_supplement', summary: '附加说明、预览或文件' },
      { name: 'qa_get_answer', summary: '阻塞等待用户回答' },
      { name: 'qa_reget_answer', summary: '重新取回已提交的回答' },
      { name: 'qa_cancel_question', summary: '取消当前问题' },
    ],
  },
  {
    id: 'preview',
    name: 'Preview',
    summary:
      '把 HTML、CSS 和 JavaScript 原型放进安全沙盒，先看见真实效果，再决定怎样落到项目里。',
    tools: [
      { name: 'preview_session_list', summary: '复用当前任务的预览会话' },
      { name: 'preview_session_create', summary: '创建空会话，不会生成代码' },
      { name: 'preview_file_upload', summary: '逐文件上传，单文件上限 256KB' },
      { name: 'preview_file_list', summary: '打开页面前核对文件清单' },
      { name: 'preview_file_get', summary: '读取单个预览文件' },
    ],
  },
  {
    id: 'pages',
    name: 'Pages',
    summary:
      '把核对完成的 Preview 晋升为项目级不可变快照，按路径对外访问；密码只在控制台设置。',
    tools: [
      { name: 'pages_list', summary: '列出项目已发布页面与生效版本' },
      { name: 'pages_promote', summary: '将预览会话晋升为 Pages 快照' },
      { name: 'pages_fork', summary: '从已发布页面派生可继续修改的 Preview 草稿' },
    ],
  },
  {
    id: 'repowiki',
    name: 'RepoWiki',
    summary:
      '随时翻阅已经生成的项目 Wiki，快速找到架构、模块和实现说明。Wiki 会随 Git Webhook 更新。',
    note: '只读',
    tools: [
      { name: 'repoWiki_list', summary: '列出已完成的 Wiki 版本' },
      { name: 'repoWiki_query', summary: '读取指定版本的页面正文' },
    ],
  },
  {
    id: 'pin',
    name: 'Pin',
    summary:
      '把依赖升级、接口变化等约束定向交给另一个项目，并按到达顺序逐条处理。',
    tools: [
      { name: 'pin_peek', summary: '预览队首约束，不消费' },
      { name: 'pin_list', summary: '列出目标项目的约束' },
      { name: 'pin_push', summary: '向另一项目推送约束' },
      { name: 'pin_consume', summary: '按 FIFO 或 ID 消费' },
      { name: 'pin_update', summary: '更新尚未消费的约束' },
    ],
  },
]

export const MCP_TOOL_NAMES: string[] = MCP_TOOL_MODULES.flatMap((module) =>
  module.tools.map((tool) => tool.name),
)

export const MCP_WORKFLOW_STEPS: McpWorkflowStep[] = [
  {
    step: 1,
    title: '先选定空间，再找到当前项目',
    body: 'Agent 先列出空间并选定一个，再按项目路径在该空间内查找。第一次接入的代码库，确认没有重复记录后再创建。',
    tools: ['workspace_list', 'workspace_get', 'project_get', 'project_list', 'project_create'],
  },
  {
    step: 2,
    title: '需要你决定时，直接来问',
    body: '方案有分支、信息不完整或需要文件时，Agent 会把结构化问题送到交互页，并在收到回答后继续工作。',
    tools: [
      'qa_session_list',
      'qa_session_create',
      'qa_what_question',
      'qa_push_question',
      'qa_push_supplement',
      'qa_get_answer',
    ],
  },
  {
    step: 3,
    title: '界面想法先放进预览',
    body: '涉及页面或交互时，可以先上传一个可操作的沙盒原型。核对文件清单后，再把预览交给你体验。',
    tools: [
      'preview_session_list',
      'preview_session_create',
      'preview_file_upload',
      'preview_file_list',
    ],
  },
  {
    step: 4,
    title: '草稿确认后晋升为 Pages',
    body: '需要对外路径式访问时，先列出已发布页面确认 slug，再把当前 Preview 晋升为不可变快照。密码保护只在控制台配置。',
    tools: ['pages_list', 'pages_promote', 'pages_fork'],
  },
  {
    step: 5,
    title: '把预览带回同一次讨论',
    body: 'Agent 可以直接打开预览链接，也可以把预览作为问答补充发到交互页，让设计和反馈留在同一个会话里。',
    tools: ['qa_push_supplement', 'qa_get_answer'],
  },
  {
    step: 6,
    title: '需要全局视野时翻阅 Wiki',
    body: '先找到已完成的 Wiki 版本，再读取相关页面。版本更新由仓库的 Git Webhook 自动触发。',
    tools: ['repoWiki_list', 'repoWiki_query'],
  },
  {
    step: 7,
    title: '跨项目变化及时传递',
    body: '接口变化或依赖升级会影响其他代码库时，先查看目标项目的待处理约束，再推送或消费对应记录。',
    tools: ['pin_peek', 'pin_list', 'pin_push', 'pin_consume'],
  },
]
