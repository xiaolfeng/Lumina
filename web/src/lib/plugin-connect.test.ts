import { describe, expect, it } from 'vitest'
import {
  MARKETPLACE_NAME,
  PLUGIN_MARKETPLACE_PATH,
  PLUGIN_NAME,
  PLUGIN_ZIP_PATH,
  buildClaudeMarketplaceAdd,
  buildClaudePluginEnv,
  buildClaudePluginInstall,
  buildClaudePluginSnippet,
  buildCodexMarketplaceAdd,
  buildCodexPluginAdd,
  buildCodexPluginSnippet,
  buildNpxSkillsAddHost,
  buildNpxSkillsAddZip,
  buildNpxSkillsSnippet,
  buildPluginMarketplaceUrl,
  buildPluginZipUrl,
} from './plugin-connect'

const origin = 'http://127.0.0.1:8800'

describe('plugin install URLs', () => {
  it('joins origin with marketplace and zip paths, stripping trailing slashes', () => {
    expect(buildPluginMarketplaceUrl(`${origin}/`)).toBe(
      `${origin}${PLUGIN_MARKETPLACE_PATH}`,
    )
    expect(buildPluginZipUrl(`${origin}/`)).toBe(`${origin}${PLUGIN_ZIP_PATH}`)
    expect(buildPluginMarketplaceUrl('')).toBe(PLUGIN_MARKETPLACE_PATH)
    expect(buildPluginZipUrl('')).toBe(PLUGIN_ZIP_PATH)
  })
})

describe('install snippets', () => {
  it('builds Claude Code marketplace add + install', () => {
    expect(buildClaudeMarketplaceAdd(origin)).toBe(
      `claude plugin marketplace add ${origin}${PLUGIN_MARKETPLACE_PATH}`,
    )
    expect(buildClaudePluginEnv('lumi_test')).toBe(
      "export LUMINA_API_KEY='lumi_test'",
    )
    expect(buildClaudePluginEnv()).toBe("export LUMINA_API_KEY='<api-key>'")
    expect(buildClaudePluginInstall()).toBe(
      `claude plugin install ${PLUGIN_NAME}@${MARKETPLACE_NAME}`,
    )
    expect(buildClaudePluginSnippet(origin, 'lumi_test')).toBe(
      [
        "export LUMINA_API_KEY='lumi_test'",
        `claude plugin marketplace add ${origin}${PLUGIN_MARKETPLACE_PATH}`,
        `claude plugin install ${PLUGIN_NAME}@${MARKETPLACE_NAME}`,
      ].join('\n'),
    )
  })

  it('builds npx skills add for zip and host discovery', () => {
    expect(buildNpxSkillsAddZip(origin)).toBe(
      `npx skills add ${origin}${PLUGIN_ZIP_PATH}`,
    )
    expect(buildNpxSkillsAddHost(origin)).toBe(`npx skills add ${origin}`)
    expect(buildNpxSkillsAddHost('')).toBe('npx skills add <origin>')
    expect(buildNpxSkillsSnippet(origin)).toBe(
      [
        `npx skills add ${origin}${PLUGIN_ZIP_PATH}`,
        `npx skills add ${origin}`,
      ].join('\n'),
    )
  })

  it('builds codex marketplace add from the project git repo', () => {
    expect(buildCodexMarketplaceAdd()).toBe(
      'codex plugin marketplace add xiaolfeng/Lumina',
    )
    expect(buildCodexPluginAdd()).toBe(
      `codex plugin add ${PLUGIN_NAME}@${MARKETPLACE_NAME}`,
    )
    expect(buildCodexPluginSnippet()).toBe(
      [
        'codex plugin marketplace add xiaolfeng/Lumina',
        `codex plugin add ${PLUGIN_NAME}@${MARKETPLACE_NAME}`,
      ].join('\n'),
    )
  })
})
