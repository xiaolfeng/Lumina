import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'
import { buildNpxSkillsSnippet } from '#/lib/plugin-connect'

interface SkillsInstallPanelProps {
  origin: string
}

export function SkillsInstallPanel({ origin }: SkillsInstallPanelProps) {
  return (
    <div>
      <ChannelKicker>方案三 · 仅安装技能</ChannelKicker>
      <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
        只装 Q&amp;A、Preview、Pin 与 RepoWiki 技能
      </h2>
      <p className="mt-2 mb-4 max-w-[48em] text-[13px] leading-relaxed text-sea-ink-soft">
        只想把 Lumina 的使用说明装进支持 Agent Skills 的客户端、不连接 MCP
        时，选择这个方式。它不会自动配置
        MCP；需要工具调用时请回到方案一或方案二。
      </p>
      {origin ? (
        <CopyBlock code={buildNpxSkillsSnippet(origin)} filename="npx skills" />
      ) : (
        <p className="text-sm text-sea-ink-soft">正在准备插件地址…</p>
      )}
    </div>
  )
}
