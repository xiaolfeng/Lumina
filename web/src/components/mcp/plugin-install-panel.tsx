import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@lumina/components/ui/tabs'
import {
  buildClaudePluginSnippet,
  buildNpxSkillsSnippet,
  buildPluginMarketplaceUrl,
  buildPluginZipUrl,
} from '#/lib/plugin-connect'
import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'

interface PluginInstallPanelProps {
  origin: string
}

export function PluginInstallPanel({ origin }: PluginInstallPanelProps) {
  const marketplaceUrl = origin ? buildPluginMarketplaceUrl(origin) : ''
  const zipUrl = origin ? buildPluginZipUrl(origin) : ''

  return (
    <div>
      <ChannelKicker>技能</ChannelKicker>
      <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
        插件与技能
      </h2>
      <p className="mt-1 mb-4 text-[13px] leading-relaxed text-sea-ink-soft">
        源码内嵌在当前二进制里，请求时即时打包。Claude Code 走市场清单；Vercel /
        通用 Agent 走 ZIP 或域名探测。
      </p>

      <dl className="mb-4 grid grid-cols-1 gap-px bg-line sm:grid-cols-2">
        <div className="bg-sand px-3.5 py-3">
          <dt className="text-[11px] font-semibold uppercase tracking-[0.14em] text-sea-ink-soft">
            市场清单
          </dt>
          <dd className="mt-1 font-mono text-[13px] leading-snug break-all text-sea-ink">
            {marketplaceUrl || '（等待解析域名）'}
          </dd>
        </div>
        <div className="bg-sand px-3.5 py-3">
          <dt className="text-[11px] font-semibold uppercase tracking-[0.14em] text-sea-ink-soft">
            插件包
          </dt>
          <dd className="mt-1 font-mono text-[13px] leading-snug break-all text-sea-ink">
            {zipUrl || '（等待解析域名）'}
          </dd>
        </div>
      </dl>

      {origin ? (
        <Tabs defaultValue="claude-code" className="gap-3">
          <TabsList
            variant="line"
            className="h-auto w-full flex-wrap justify-start gap-0 rounded-none bg-transparent p-0"
          >
            <TabsTrigger
              value="claude-code"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              Claude Code
            </TabsTrigger>
            <TabsTrigger
              value="vercel"
              className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
            >
              Vercel / 通用
            </TabsTrigger>
          </TabsList>
          <TabsContent value="claude-code" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              <span className="font-medium text-sea-ink">终端命令</span>
              <span className="mx-1.5 text-line">·</span>
              先加入本站市场，再安装 lumina@lumina。本地 HTTP
              地址在部分客户端上可能被拒绝，正式环境请用 HTTPS 域名。
            </p>
            <CopyBlock
              code={buildClaudePluginSnippet(origin)}
              filename="Claude Code"
            />
          </TabsContent>
          <TabsContent value="vercel" className="space-y-3">
            <p className="text-[13px] leading-relaxed text-sea-ink-soft">
              <span className="font-medium text-sea-ink">npx skills</span>
              <span className="mx-1.5 text-line">·</span>
              可直接下载 ZIP，或把站点根地址交给域名探测协议。
            </p>
            <CopyBlock
              code={buildNpxSkillsSnippet(origin)}
              filename="npx skills"
            />
          </TabsContent>
        </Tabs>
      ) : (
        <p className="text-sm text-sea-ink-soft">正在解析接入地址…</p>
      )}
    </div>
  )
}
