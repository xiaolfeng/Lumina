import { describe, expect, it } from 'vitest'
import {
  MCP_CLIENTS,
  MCP_KEY_PLACEHOLDER,
  MCP_PATH,
  MCP_SERVER_NAME,
  MCP_TOOL_MODULES,
  MCP_TOOL_NAMES,
  MCP_WORKFLOW_STEPS,
  buildAuthorizationHeader,
  buildAuthorizationValue,
  buildClaudeOAuthAdd,
  buildClientSnippet,
  buildCodexOAuthAdd,
  buildGenericOAuthJSON,
  buildMcpUrl,
  normalizeApiKey,
  resolveMcpOrigin,
} from './mcp-connect'

const ctx = {
  mcpUrl: 'https://lumina.example.com/api/v1/mcp',
  apiKey: 'lumi_testkey',
}

const expectedTools = [
  'qa_session_create',
  'qa_session_list',
  'qa_session_get',
  'qa_session_archive',
  'qa_push_question',
  'qa_push_supplement',
  'qa_what_question',
  'qa_get_answer',
  'qa_reget_answer',
  'qa_cancel_question',
  'project_create',
  'project_get',
  'project_list',
  'pin_push',
  'pin_consume',
  'pin_list',
  'pin_update',
  'pin_peek',
  'repoWiki_query',
  'repoWiki_list',
  'preview_session_create',
  'preview_session_list',
  'preview_file_upload',
  'preview_file_edit',
  'preview_file_delete',
  'preview_file_list',
  'preview_file_get',
  'preview_lpw_init',
  'preview_lpw_node_add',
  'preview_lpw_node_edit',
  'preview_lpw_node_remove',
  'preview_lpw_node_sort',
  'preview_lpw_meta_set',
  'preview_lpw_outline',
  'preview_lpw_schema',
  'workspace_list',
  'workspace_get',
  'pages_list',
  'pages_promote',
  'pages_fork',
]

describe('resolveMcpOrigin', () => {
  it('prefers override, then site.domain, then window origin', () => {
    expect(
      resolveMcpOrigin({
        override: 'https://override.example/',
        siteDomain: 'https://site.example',
        windowOrigin: 'http://localhost:3000',
      }),
    ).toBe('https://override.example')

    expect(
      resolveMcpOrigin({
        override: '  ',
        siteDomain: 'https://site.example/',
        windowOrigin: 'http://localhost:3000',
      }),
    ).toBe('https://site.example')

    expect(
      resolveMcpOrigin({
        siteDomain: '',
        windowOrigin: 'http://localhost:3000/',
      }),
    ).toBe('http://localhost:3000')
  })
})

describe('buildMcpUrl', () => {
  it('joins origin with /api/v1/mcp and strips trailing slashes', () => {
    expect(buildMcpUrl('http://127.0.0.1:8800/')).toBe(
      `http://127.0.0.1:8800${MCP_PATH}`,
    )
    expect(buildMcpUrl('')).toBe(MCP_PATH)
  })
})

describe('api key helpers', () => {
  it('falls back to placeholder and builds Bearer header', () => {
    expect(normalizeApiKey(undefined)).toBe(MCP_KEY_PLACEHOLDER)
    expect(normalizeApiKey('  ')).toBe(MCP_KEY_PLACEHOLDER)
    expect(normalizeApiKey(' lumi_abc ')).toBe('lumi_abc')
    expect(MCP_KEY_PLACEHOLDER).toBe('<api-key>')
    expect(buildAuthorizationValue()).toBe(`Bearer ${MCP_KEY_PLACEHOLDER}`)
    expect(buildAuthorizationHeader('lumi_abc')).toBe(
      'Authorization: Bearer lumi_abc',
    )
  })
})

describe('tool catalog', () => {
  it('lists the 40 backend MCP tools', () => {
    expect(MCP_TOOL_NAMES).toHaveLength(40)
    expect([...MCP_TOOL_NAMES].sort()).toEqual([...expectedTools].sort())
    expect(MCP_TOOL_NAMES).not.toContain('qa_pushQuestion')
    expect(MCP_TOOL_NAMES).not.toContain('memory_create')
    expect(MCP_TOOL_NAMES).not.toContain('repoWiki_analyze')
  })

  it('organizes tools into 8 core capability modules', () => {
    expect(MCP_TOOL_MODULES).toHaveLength(8)
    const moduleIds = MCP_TOOL_MODULES.map((mod) => mod.id)
    expect(moduleIds).toEqual([
      'workspace',
      'project',
      'qa',
      'preview',
      'preview_lpw',
      'pages',
      'repowiki',
      'pin',
    ])

    const previewMod = MCP_TOOL_MODULES.find((m) => m.id === 'preview')
    expect(previewMod?.tools.map((t) => t.name)).toContain('preview_file_edit')
    expect(previewMod?.tools.map((t) => t.name)).toContain(
      'preview_file_delete',
    )
    expect(previewMod?.tools).toHaveLength(7)

    const lpwMod = MCP_TOOL_MODULES.find((m) => m.id === 'preview_lpw')
    expect(lpwMod?.tools).toHaveLength(8)
    expect(lpwMod?.tools.map((t) => t.name)).toEqual([
      'preview_lpw_init',
      'preview_lpw_node_add',
      'preview_lpw_node_edit',
      'preview_lpw_node_remove',
      'preview_lpw_node_sort',
      'preview_lpw_meta_set',
      'preview_lpw_outline',
      'preview_lpw_schema',
    ])
  })

  it('includes incremental editing and LPW steps in workflow', () => {
    expect(MCP_WORKFLOW_STEPS).toHaveLength(8)
    const step3 = MCP_WORKFLOW_STEPS.find((s) => s.step === 3)
    expect(step3?.tools).toContain('preview_file_edit')
    expect(step3?.tools).toContain('preview_file_delete')

    const step4 = MCP_WORKFLOW_STEPS.find((s) => s.step === 4)
    expect(step4?.tools).toContain('preview_lpw_init')
    expect(step4?.tools).toContain('preview_lpw_node_add')
  })
})

describe('client snippets', () => {
  it('registers the documented clients', () => {
    expect(MCP_CLIENTS.map((client) => client.id)).toEqual([
      'cursor',
      'windsurf',
      'vscode',
      'cline',
      'claude-desktop',
      'claude-code',
      'codex',
      'grok',
      'generic',
    ])
  })

  it('builds Cursor url + Authorization header JSON', () => {
    const parsed = JSON.parse(buildClientSnippet('cursor', ctx))
    expect(parsed.mcpServers.lumina.url).toBe(ctx.mcpUrl)
    expect(parsed.mcpServers.lumina.headers.Authorization).toBe(
      'Bearer lumi_testkey',
    )
  })

  it('builds Windsurf serverUrl JSON', () => {
    const parsed = JSON.parse(buildClientSnippet('windsurf', ctx))
    expect(parsed.mcpServers.lumina.serverUrl).toBe(ctx.mcpUrl)
    expect(parsed.mcpServers.lumina.url).toBeUndefined()
  })

  it('builds VS Code servers.http JSON', () => {
    const parsed = JSON.parse(buildClientSnippet('vscode', ctx))
    expect(parsed.servers.lumina).toEqual({
      type: 'http',
      url: ctx.mcpUrl,
      headers: { Authorization: 'Bearer lumi_testkey' },
    })
  })

  it('builds Cline streamableHttp JSON', () => {
    const parsed = JSON.parse(buildClientSnippet('cline', ctx))
    expect(parsed.mcpServers.lumina.type).toBe('streamableHttp')
    expect(parsed.mcpServers.lumina.url).toBe(ctx.mcpUrl)
  })

  it('bridges Claude Desktop via mcp-remote without spaces in args header', () => {
    const httpsParsed = JSON.parse(buildClientSnippet('claude-desktop', ctx))
    expect(httpsParsed.mcpServers.lumina.command).toBe('npx')
    expect(httpsParsed.mcpServers.lumina.args).toEqual([
      '-y',
      'mcp-remote',
      ctx.mcpUrl,
      '--header',
      'Authorization:${AUTH_HEADER}',
    ])
    expect(httpsParsed.mcpServers.lumina.env.AUTH_HEADER).toBe(
      'Bearer lumi_testkey',
    )
    expect(httpsParsed.mcpServers.lumina.args).not.toContain('--allow-http')

    const httpParsed = JSON.parse(
      buildClientSnippet('claude-desktop', {
        mcpUrl: 'http://localhost:8080/api/v1/mcp',
        apiKey: 'lumi_testkey',
      }),
    )
    expect(httpParsed.mcpServers.lumina.args).toContain('--allow-http')
  })

  it('builds Claude Code CLI with Bearer header', () => {
    expect(buildClientSnippet('claude-code', ctx)).toBe(
      `claude mcp add --transport http lumina ${ctx.mcpUrl} --header "Authorization: Bearer lumi_testkey"`,
    )
  })

  it('builds Codex TOML with mcp_servers and http_headers', () => {
    const snippet = buildClientSnippet('codex', ctx)
    expect(snippet).toContain('[mcp_servers.lumina]')
    expect(snippet).toContain(`url = "${ctx.mcpUrl}"`)
    expect(snippet).toContain(
      'http_headers = { Authorization = "Bearer lumi_testkey" }',
    )
  })

  it('builds Grok TOML with headers', () => {
    const snippet = buildClientSnippet('grok', ctx)
    expect(snippet).toContain('[mcp_servers.lumina]')
    expect(snippet).toContain(
      'headers = { Authorization = "Bearer lumi_testkey" }',
    )
  })

  it('uses placeholder when api key is missing', () => {
    const parsed = JSON.parse(
      buildClientSnippet('cursor', { mcpUrl: ctx.mcpUrl }),
    )
    expect(parsed.mcpServers.lumina.headers.Authorization).toBe(
      `Bearer ${MCP_KEY_PLACEHOLDER}`,
    )
  })

  it('describes generic Streamable HTTP without SSE server', () => {
    const snippet = buildClientSnippet('generic', ctx)
    expect(snippet).toContain('Streamable HTTP')
    expect(snippet).toContain(ctx.mcpUrl)
    expect(snippet).toContain('Authorization: Bearer lumi_testkey')
    expect(snippet).not.toMatch(/独立.*SSE/)
  })
})

describe('oauth direct-connect snippets', () => {
  const mcpUrl = 'https://lumina.example.com/api/v1/mcp'

  it('adds Claude Code server without auth header', () => {
    expect(buildClaudeOAuthAdd(mcpUrl)).toBe(
      `claude mcp add --transport http ${MCP_SERVER_NAME} ${mcpUrl}`,
    )
    expect(buildClaudeOAuthAdd(mcpUrl)).not.toContain('--header')
  })

  it('adds Codex server via url flag', () => {
    expect(buildCodexOAuthAdd(mcpUrl)).toBe(
      `codex mcp add ${MCP_SERVER_NAME} --url ${mcpUrl}`,
    )
  })

  it('builds generic mcpServers JSON without Authorization', () => {
    const parsed = JSON.parse(buildGenericOAuthJSON(mcpUrl))
    expect(parsed.mcpServers[MCP_SERVER_NAME]).toEqual({
      type: 'http',
      url: mcpUrl,
    })
    expect(buildGenericOAuthJSON(mcpUrl)).not.toContain('Authorization')
  })
})
