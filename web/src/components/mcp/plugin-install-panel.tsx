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
  buildClaudeRepoMarketplaceAdd,
  buildPluginMarketplaceUrl,
  buildPluginMarketplaceZcodeUrl,
} from '#/lib/plugin-connect'
import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'

interface PluginInstallPanelProps {
  origin: string
}

const pluginBenefits = [
  '自动连接当前 Lumina 的 25 个 MCP 工具',
  '同时安装 Q&A、Preview、Pin 与 RepoWiki 技能',
  '后续更新继续沿用同一个插件入口',
]

export function PluginInstallPanel({ origin }: PluginInstallPanelProps) {
  const marketplaceUrl = origin ? buildPluginMarketplaceUrl(origin) : ''
  const zcodeMarketplaceUrl = origin
    ? buildPluginMarketplaceZcodeUrl(origin)
    : ''

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <ChannelKicker>方案一 · 最优</ChannelKicker>
          <h1 className="display-title mt-2 text-[30px] font-medium tracking-tight text-sea-ink sm:text-[34px]">
            安装插件，一步到位
          </h1>
        </div>
        <span className="inline-flex items-center gap-1.5 bg-lagoon px-3 py-1.5 text-[11px] font-bold tracking-[0.12em] text-foam">
          <Sparkles className="size-3.5" aria-hidden />
          完整技能包
        </span>
      </div>

      <p className="mt-3 max-w-[48em] text-[15px] leading-relaxed text-sea-ink-soft">
        插件随包携带 Q&A、Preview、Pin 与 RepoWiki 技能说明，并按客户端机制读取随包提供的
        MCP 配置连接到当前站点。MCP 采用 OAuth 登录授权，首次连接时完成一次登录即可，无需手动管理密钥。
      </p>

      <div className="mt-5 grid gap-px bg-line sm:grid-cols-1">
        <div className="bg-sand p-4 sm:p-5">
          <div className="flex items-center gap-2 text-sea-ink">
            <PackageOpen className="size-4 text-lagoon-deep" aria-hidden />
            <h2 className="text-[15px] font-semibold">
              插件会替你准备好这些内容
            </h2>
          </div>
          <ul className="mt-3 grid gap-2.5 sm:grid-cols-3">
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
          </TabsList>
          <TabsContent value="claude-code" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              复制并运行下面两行命令。安装完成后，重启 Claude Code，再用{' '}
              <code>/mcp</code> 查看 Lumina 是否已连接；首次连接时按提示完成
              OAuth 登录授权。
            </p>
            <CopyBlock
              code={buildClaudePluginSnippet(origin)}
              filename="Claude Code · 推荐"
            />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              插件包来自{' '}
              <span className="font-mono break-all text-sea-ink">
                {marketplaceUrl}
              </span>
              ，其中的 MCP 地址会指向当前 Lumina 站点。
            </p>
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              <Wrench className="mr-1 inline size-3.5" aria-hidden />
              旧版本 Claude Code 不支持 archive
              插件源时，可改用仓库市场：
              <span className="font-mono break-all text-sea-ink">
                {buildClaudeRepoMarketplaceAdd()}
              </span>
              （仅含技能，MCP 需按方案二手动接入）。
            </p>
          </TabsContent>
          <TabsContent value="zcode" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              打开 ZCode 的「设置 → 插件管理 → 发现」，点击{' '}
              <code>+</code>{' '}
              粘贴下方专用市场地址并安装 <code>lumina</code>{' '}
              插件，完成后重启 ZCode。
            </p>
            <CopyBlock
              code={zcodeMarketplaceUrl}
              filename="ZCode · 专用市场地址"
            />
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              该地址始终返回 ZCode 兼容的 url + zip
              清单变体，不会出现「不支持 archive
              源」的报错。若 ZCode 未自动完成 MCP 连接，请使用方案二手动接入。
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
              技能。Codex 不消费动态插件包内的 MCP 配置，连接 MCP 请使用方案二的
              OAuth 直连（<code>codex mcp login</code>）或 API Key 配置。
            </p>
          </TabsContent>
        </Tabs>
      ) : (
        <p className="mt-5 text-sm text-sea-ink-soft">正在准备插件地址…</p>
      )}
    </div>
  )
}
