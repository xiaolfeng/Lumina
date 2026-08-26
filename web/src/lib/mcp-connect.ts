/**
 * MCP 接入配置的单一数据源。
 * 工具名必须与 internal/mcp/{qa,project,pin,repowiki,preview}_tools.go 保持一致。
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
    note: '执行后可用 claude mcp list 确认。',
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
    note: 'Q&A 实时通道是 WebSocket，与 MCP 传输无关。不要再配独立的 SSE 服务。',
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
    id: 'project',
    name: 'Project',
    summary: '把代码库注册到微明，后续问答和预览都挂在 project_id 上。',
    tools: [
      { name: 'project_get', summary: '按 ID、名称或路径解析已有项目' },
      { name: 'project_list', summary: '列出项目，可用当前工作目录过滤' },
      { name: 'project_create', summary: '仅在确认尚未注册且任务需要时创建' },
    ],
  },
  {
    id: 'qa',
    name: 'Q&A',
    summary: '向用户提问并等待回答。实时推送走 WebSocket，不是 SSE。',
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
    summary: '把 HTML/CSS/JS 原型推给用户评审，不能替代改真实仓库。',
    tools: [
      { name: 'preview_session_list', summary: '复用当前任务的预览会话' },
      { name: 'preview_session_create', summary: '创建空会话，不会生成代码' },
      { name: 'preview_file_upload', summary: '逐文件上传，单文件上限 256KB' },
      { name: 'preview_file_list', summary: '打开页面前核对文件清单' },
      { name: 'preview_file_get', summary: '读取单个预览文件' },
    ],
  },
  {
    id: 'repowiki',
    name: 'RepoWiki',
    summary:
      '只读查询已生成的 Wiki。更新由 Git Webhook 触发，MCP 不能分析仓库。',
    note: '只读',
    tools: [
      { name: 'repoWiki_list', summary: '列出已完成的 Wiki 版本' },
      { name: 'repoWiki_query', summary: '读取指定版本的页面正文' },
    ],
  },
  {
    id: 'pin',
    name: 'Pin',
    summary: '跨项目约束的点对点推送与 FIFO 消费，不要拿来当普通备忘。',
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
    title: '解析项目',
    body: '先用 project_get（优先 match_path）或 project_list 找到 project_id。确认尚未注册且任务确实需要时，才 project_create。',
    tools: ['project_get', 'project_list', 'project_create'],
  },
  {
    step: 2,
    title: '向用户提问',
    body: 'qa_session_list / create → qa_what_question → qa_push_question。需要补充时先 qa_push_supplement，再 qa_get_answer。不要用猜测代替用户决策。',
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
    title: '可视化评审',
    body: 'preview_session_list / create 后逐个 preview_file_upload，最后 preview_file_list 核对。Preview 是沟通媒介，不能代替改真实项目文件。',
    tools: [
      'preview_session_list',
      'preview_session_create',
      'preview_file_upload',
      'preview_file_list',
    ],
  },
  {
    step: 4,
    title: '把预览交给用户',
    body: '优先打开返回的绝对 preview_url。作为问答补充时，qa_push_supplement 的 content 必须原样使用 Preview 返回的 qa_supplement.content。hash 只用于网页 URL。',
    tools: ['qa_push_supplement', 'qa_get_answer'],
  },
  {
    step: 5,
    title: '查阅 Wiki',
    body: 'RepoWiki MCP 只读。用 repoWiki_list 找完成版本，再用 repoWiki_query 读页面。仓库分析由 Git Webhook 驱动。',
    tools: ['repoWiki_list', 'repoWiki_query'],
  },
  {
    step: 6,
    title: '跨项目约束',
    body: 'Pin 只用于明确的跨项目约束。消费前可先 pin_peek / pin_list 核对，再 pin_push 或 pin_consume。',
    tools: ['pin_peek', 'pin_list', 'pin_push', 'pin_consume'],
  },
]
