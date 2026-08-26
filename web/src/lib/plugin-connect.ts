/**
 * AI 插件 / 技能安装命令的单一数据源。
 * 路径必须与 internal/constant/ai_plugin.go 以及 route_plugin.go 保持一致。
 */

import { trimSlash } from './mcp-connect'

export const PLUGIN_MARKETPLACE_PATH = '/api/v1/plugins/marketplace.json'
export const PLUGIN_ZIP_PATH = '/api/v1/plugins/lumina.zip'
export const PLUGIN_NAME = 'lumina'
export const MARKETPLACE_NAME = 'lumina'

export function buildPluginMarketplaceUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return PLUGIN_MARKETPLACE_PATH
  return `${base}${PLUGIN_MARKETPLACE_PATH}`
}

export function buildPluginZipUrl(origin: string): string {
  const base = trimSlash(origin)
  if (!base) return PLUGIN_ZIP_PATH
  return `${base}${PLUGIN_ZIP_PATH}`
}

export function buildClaudeMarketplaceAdd(origin: string): string {
  return `claude plugin marketplace add ${buildPluginMarketplaceUrl(origin)}`
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

export function buildClaudePluginSnippet(origin: string): string {
  return [buildClaudeMarketplaceAdd(origin), buildClaudePluginInstall()].join(
    '\n',
  )
}

export function buildNpxSkillsSnippet(origin: string): string {
  return [buildNpxSkillsAddZip(origin), buildNpxSkillsAddHost(origin)].join(
    '\n',
  )
}
