import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { ChannelBand, ChannelKicker } from '#/components/mcp/channel-band'
import { ClientConfigPanel } from '#/components/mcp/client-config-panel'
import { EndpointCard } from '#/components/mcp/endpoint-card'
import { OAuthConnectPanel } from '#/components/mcp/oauth-connect-panel'
import { PluginInstallPanel } from '#/components/mcp/plugin-install-panel'
import { ToolCatalog, WorkflowList } from '#/components/mcp/tool-catalog'
import { useMcpEndpoint } from '#/hooks/useMcpEndpoint'

export const Route = createFileRoute('/console/connect')({
  staticData: { crumb: '接入指南' },
  component: ConnectPage,
})

function ConnectPage() {
  const [apiKey, setApiKey] = useState('')
  const { mcpUrl, origin, source } = useMcpEndpoint()

  return (
    <motion.div
      className="-mx-4 max-w-5xl overflow-hidden border-y border-line sm:mx-auto"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <motion.div variants={staggerItem}>
        <ChannelBand index={1}>
          {mcpUrl ? (
            <OAuthConnectPanel mcpUrl={mcpUrl} />
          ) : (
            <p className="text-sm text-sea-ink-soft">正在准备 MCP 地址…</p>
          )}
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={2} alt>
          <PluginInstallPanel origin={origin} apiKey={apiKey} />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={3}>
          <ChannelKicker>方案 C · 手动接入</ChannelKicker>
          <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
            API Key 独立配置 MCP
          </h2>
          <p className="mt-2 mb-4 max-w-[48em] text-[13px] leading-relaxed text-sea-ink-soft">
            不方便走 OAuth 登录的环境，可以先在本页生成 API
            Key，再把带鉴权头的配置片段写入 Cursor、Windsurf、VS Code、Codex、Grok
            等客户端。Claude Code
            也保留手动接入方式，适合不安装插件的环境。
          </p>
          <EndpointCard
            mcpUrl={mcpUrl}
            origin={origin}
            source={source}
            apiKey={apiKey}
            onApiKeyChange={setApiKey}
          />
          {mcpUrl ? (
            <div className="mt-4">
              <ClientConfigPanel mcpUrl={mcpUrl} apiKey={apiKey} />
            </div>
          ) : (
            <p className="text-sm text-sea-ink-soft">正在准备 MCP 地址…</p>
          )}
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={4} alt>
          <ToolCatalog />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={5}>
          <WorkflowList />
        </ChannelBand>
      </motion.div>
    </motion.div>
  )
}
