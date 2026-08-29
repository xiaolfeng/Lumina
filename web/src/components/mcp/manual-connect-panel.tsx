import { useState } from 'react'
import { KeyRound, PlugZap, ShieldCheck } from 'lucide-react'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@lumina/components/ui/tabs'

import { ChannelKicker } from './channel-band'
import { ClientConfigPanel } from './client-config-panel'
import { CopyBlock } from './copy-block'
import { EndpointCard } from './endpoint-card'
import {
  buildClaudeOAuthAdd,
  buildCodexOAuthAdd,
  buildGenericOAuthJSON,
} from '#/lib/mcp-connect'
import type { McpOriginSource } from '#/hooks/useMcpEndpoint'

interface ManualConnectPanelProps {
  mcpUrl: string
  origin: string
  source: McpOriginSource
}

export function ManualConnectPanel({
  mcpUrl,
  origin,
  source,
}: ManualConnectPanelProps) {
  const [apiKey, setApiKey] = useState('')

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <ChannelKicker>方案二 · 手动接入</ChannelKicker>
          <h2 className="display-title mt-2 text-[26px] font-medium tracking-tight text-sea-ink sm:text-[30px]">
            手动安装 MCP
          </h2>
        </div>
      </div>

      <p className="mt-3 max-w-[48em] text-[15px] leading-relaxed text-sea-ink-soft">
        不安装插件、或客户端未被插件覆盖时，可以手动添加 MCP
        地址。鉴权方式二选一：推荐 OAuth2
        登录授权，令牌自动签发与续期；也可以使用长期 API Key。
      </p>

      <Tabs defaultValue="oauth" className="mt-5 gap-3">
        <TabsList
          variant="line"
          className="h-auto w-full flex-wrap justify-start gap-0 rounded-none bg-transparent p-0"
        >
          <TabsTrigger
            value="oauth"
            className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
          >
            <ShieldCheck className="mr-1.5 inline size-3.5" aria-hidden />
            OAuth2 登录
          </TabsTrigger>
          <TabsTrigger
            value="apikey"
            className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
          >
            <KeyRound className="mr-1.5 inline size-3.5" aria-hidden />
            API Key
          </TabsTrigger>
        </TabsList>

        <TabsContent value="oauth" className="space-y-3">
          <p className="text-[13px] leading-relaxed text-sea-ink-soft">
            <PlugZap
              className="mr-1 inline size-3.5 text-lagoon-deep"
              aria-hidden
            />
            添加不带鉴权头的 MCP 地址即可，客户端首次连接会打开 Lumina
            授权页，登录确认后令牌自动签发并续期。
          </p>
          <CopyBlock code={mcpUrl} filename="MCP 端点" />
          <CopyBlock
            code={buildClaudeOAuthAdd(mcpUrl)}
            filename="Claude Code · 之后在 /mcp 中 Authenticate"
          />
          <CopyBlock
            code={buildCodexOAuthAdd(mcpUrl)}
            filename="Codex · 之后执行 codex mcp login lumina"
          />
          <CopyBlock
            code={buildGenericOAuthJSON(mcpUrl)}
            filename="ZCode / Cursor / 通用 · mcpServers 配置"
          />
          <p className="text-xs leading-relaxed text-sea-ink-soft">
            通用配置写入客户端 MCP 配置文件后重启客户端并触发连接，即可看到
            Lumina 授权页。
          </p>
        </TabsContent>

        <TabsContent value="apikey" className="space-y-4">
          <EndpointCard
            mcpUrl={mcpUrl}
            origin={origin}
            source={source}
            apiKey={apiKey}
            onApiKeyChange={setApiKey}
          />
          <ClientConfigPanel mcpUrl={mcpUrl} apiKey={apiKey} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
