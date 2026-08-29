/**
 * AI 插件 / 技能安装命令的单一数据源。
 * 路径必须与 internal/constant/ai_plugin.go 以及 route_plugin.go 保持一致。
 */

import { trimSlash } from './mcp-connect'

export const PLUGIN_MARKETPLACE_PATH = '/api/v1/plugins/marketplace.json'
export const PLUGIN_MARKETPLACE_ZCODE_PATH =
  '/api/v1/plugins/marketplace.zcode.json'
export const PLUGIN_ZIP_PATH = '/api/v1/plugins/lumina.zip'
export const PLUGIN_NAME = 'lumina'
export const MARKETPLACE_NAME = 'lumina'
/** Codex 市场只能来自 Git 仓库，清单内置于仓库 .agents/plugins/marketplace.json。 */
export const CODEX_MARKETPLACE_REPO = 'xiaolfeng/Lumina'

export function buildPluginMarketplaceUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return PLUGIN_MARKETPLACE_PATH
  return `${base}${PLUGIN_MARKETPLACE_PATH}`
}

/** ZCode 专用清单地址：无论 User-Agent 如何，始终返回 url+zip 兼容形态 */
export function buildPluginMarketplaceZcodeUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return PLUGIN_MARKETPLACE_ZCODE_PATH
  return `${base}${PLUGIN_MARKETPLACE_ZCODE_PATH}`
}

export function buildPluginZipUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return PLUGIN_ZIP_PATH
  return `${base}${PLUGIN_ZIP_PATH}`
}

export function buildClaudeMarketplaceAdd(origin: string): string {
  return `claude plugin marketplace add ${buildPluginMarketplaceUrl(origin)}`
}

/** 仓库市场兜底：Claude Code 版本不支持 archive 源（<2.1.224）时使用 */
export function buildClaudeRepoMarketplaceAdd(): string {
  return `claude plugin marketplace add ${CODEX_MARKETPLACE_REPO}`
}

export function buildClaudePluginInstall(): string {
  return `claude plugin install ${PLUGIN_NAME}@${MARKETPLACE_NAME}`
}

export function buildNpxSkillsAddZip(origin: string): string {
  return `npx skills add ${buildPluginZipUrl(origin)}`
}

export function buildNpxSkillsAddHost(origin: string): string {
  const base = trimSlash(origin)
  return base ? `npx skills add ${base}` : 'npx skills add <origin>'
}

export function buildNpxSkillsSnippet(origin: string): string {
  return [buildNpxSkillsAddZip(origin), buildNpxSkillsAddHost(origin)].join(
    '\n',
  )
}

export function buildClaudePluginSnippet(origin: string): string {
  return [
    buildClaudeMarketplaceAdd(origin),
    buildClaudePluginInstall(),
  ].join('\n')
}

export function buildCodexMarketplaceAdd(): string {
  return `codex plugin marketplace add ${CODEX_MARKETPLACE_REPO}`
}

export function buildCodexPluginAdd(): string {
  return `codex plugin add ${PLUGIN_NAME}@${MARKETPLACE_NAME}`
}

export function buildCodexPluginSnippet(): string {
  return [buildCodexMarketplaceAdd(), buildCodexPluginAdd()].join('\n')
}
