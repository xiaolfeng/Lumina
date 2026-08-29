import { CheckCircle2, PackageOpen, Sparkles, Wrench } from 'lucide-react'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@lumina/components/ui/tabs'
import {
  buildClaudePluginSnippet,
  buildCodexPluginSnippet,
  buildNpxSkillsSnippet,
  buildPluginMarketplaceUrl,
  buildPluginZipUrl,
} from '#/lib/plugin-connect'
import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'

interface PluginInstallPanelProps {
  origin: string
  apiKey?: string | null
}

const pluginBenefits = [
  '自动连接当前 Lumina 的 25 个 MCP 工具',
  '同时安装 Q&A、Preview、Pin 与 RepoWiki 技能',
  '后续更新继续沿用同一个插件入口',
]

export function PluginInstallPanel({
  origin,
  apiKey,
}: PluginInstallPanelProps) {
  const marketplaceUrl = origin ? buildPluginMarketplaceUrl(origin) : ''
  const zipUrl = origin ? buildPluginZipUrl(origin) : ''

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <ChannelKicker>方案 B · 插件安装</ChannelKicker>
          <h1 className="display-title mt-2 text-[30px] font-medium tracking-tight text-sea-ink sm:text-[34px]">
            安装插件，附加技能包
          </h1>
        </div>
        <span className="inline-flex items-center gap-1.5 bg-lagoon px-3 py-1.5 text-[11px] font-bold tracking-[0.12em] text-foam">
          <Sparkles className="size-3.5" aria-hidden />
          完整技能包
        </span>
      </div>

      <p className="mt-3 max-w-[48em] text-[15px] leading-relaxed text-sea-ink-soft">
        相比方案 A
        的纯直连，插件会随包安装 Q&A、Preview、Pin 与 RepoWiki
        技能说明，并按客户端机制读取随包提供的 MCP 配置连接到当前站点；无需再执行{' '}
        <code>claude mcp add</code>，也不用手动编辑 MCP 配置文件。
      </p>

      <div className="mt-5 grid gap-px bg-line sm:grid-cols-[1.05fr_0.95fr]">
        <div className="bg-sand p-4 sm:p-5">
          <div className="flex items-center gap-2 text-sea-ink">
            <PackageOpen className="size-4 text-lagoon-deep" aria-hidden />
            <h2 className="text-[15px] font-semibold">
              插件会替你准备好这些内容
            </h2>
          </div>
          <ul className="mt-3 space-y-2.5">
            {pluginBenefits.map((benefit) => (
              <li
                key={benefit}
                className="flex gap-2.5 text-[13px] leading-relaxed text-sea-ink-soft"
              >
                <CheckCircle2
                  className="mt-0.5 size-4 shrink-0 text-lagoon-deep"
                  aria-hidden
                />
                {benefit}
              </li>
            ))}
          </ul>
        </div>
        <div className="bg-chip-bg p-4 sm:p-5">
          <div className="flex items-center gap-2 text-sea-ink">
            <Wrench className="size-4 text-lagoon-deep" aria-hidden />
            <h2 className="text-[15px] font-semibold">安装前准备</h2>
          </div>
          <p className="mt-3 text-[13px] leading-relaxed text-sea-ink-soft">
            先在本页生成令牌。命令会把它写入当前终端的{' '}
            <code>LUMINA_API_KEY</code>{' '}
            环境变量，插件用该变量完成鉴权。若需要长期使用，请把变量保存到你的
            shell 配置中。
          </p>
        </div>
      </div>

      {origin ? (
        <Tabs defaultValue="claude-code" className="mt-5 gap-3">
          <TabsList
            variant="line"
            className="h-auto w-full flex-wrap justify-start gap-0 rounded-none bg-transparent p-0"
          >
            <TabsTrigger
              value="claude-code"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              Claude Code 插件
            </TabsTrigger>
            <TabsTrigger
              value="zcode"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              ZCode 插件
            </TabsTrigger>
            <TabsTrigger
              value="codex"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              Codex 插件
            </TabsTrigger>
            <TabsTrigger
              value="skills"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              仅安装技能
            </TabsTrigger>
          </TabsList>
          <TabsContent value="claude-code" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              复制并运行下面三行命令。安装完成后，重启 Claude Code，再用{' '}
              <code>/mcp</code> 查看 Lumina 是否已连接。
            </p>
            <CopyBlock
              code={buildClaudePluginSnippet(origin, apiKey)}
              filename="Claude Code · 推荐"
            />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              插件包来自{' '}
              <span className="font-mono break-all text-sea-ink">
                {marketplaceUrl}
              </span>
              ，其中的 MCP 地址会指向当前 Lumina 站点。
            </p>
          </TabsContent>
          <TabsContent value="zcode" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              打开 ZCode 的「设置 → 插件管理 → 发现」，点击{' '}
              <code>+</code>{' '}
              粘贴下方市场地址并安装 <code>lumina</code>{' '}
              插件，完成后重启 ZCode。
            </p>
            <CopyBlock code={marketplaceUrl} filename="ZCode · 市场地址" />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              ZCode 会拿到与其兼容的 url + zip
              清单变体，其余客户端拿到 archive 变体，地址无需区分。鉴权沿用{' '}
              <code>LUMINA_API_KEY</code>{' '}
              环境变量；若 MCP 未自动连接，使用方案 C 的独立 MCP 配置手动添加。
            </p>
          </TabsContent>
          <TabsContent value="codex" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              Codex 的市场只能来自 Git
              仓库，不支持直接粘贴清单地址。复制运行下面两行命令，从项目仓库内置的市场清单安装。
            </p>
            <CopyBlock code={buildCodexPluginSnippet()} filename="Codex" />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              插件提供 Q&A、Preview、Pin 与 RepoWiki
              技能。Codex 不消费动态插件包内的 MCP 配置，连接 MCP 请使用方案 C
              的独立 MCP 配置，或改用方案 A 的 OAuth 直连。
            </p>
          </TabsContent>
          <TabsContent value="skills" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              只想把 Lumina 的使用说明装进支持 Agent Skills
              的客户端时，可以选择这个方式。它不会自动连接
              MCP，仍需使用下方的独立 MCP 配置。
            </p>
            <CopyBlock
              code={buildNpxSkillsSnippet(origin)}
              filename="npx skills"
            />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              下载地址：{' '}
              <span className="font-mono break-all text-sea-ink">{zipUrl}</span>
            </p>
          </TabsContent>
        </Tabs>
      ) : (
        <p className="mt-5 text-sm text-sea-ink-soft">正在准备插件地址…</p>
      )}
    </div>
  )
}
