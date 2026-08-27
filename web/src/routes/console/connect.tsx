import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { ChannelBand, ChannelKicker } from '#/components/mcp/channel-band'
import { ClientConfigPanel } from '#/components/mcp/client-config-panel'
import { EndpointCard } from '#/components/mcp/endpoint-card'
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
          <PluginInstallPanel origin={origin} apiKey={apiKey} />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={2} alt>
          <EndpointCard
            mcpUrl={mcpUrl}
            origin={origin}
            source={source}
            apiKey={apiKey}
            onApiKeyChange={setApiKey}
          />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={3}>
          <ChannelKicker>其他方案</ChannelKicker>
          <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
            独立配置 MCP
          </h2>
          <p className="mt-2 mb-4 max-w-[48em] text-[13px] leading-relaxed text-sea-ink-soft">
            Cursor、Windsurf、VS Code、Codex、Grok 与其他支持 Streamable HTTP
            的客户端，可以直接连接
            Lumina。选择你的客户端，复制对应片段即可。Claude Code
            也保留手动接入方式，适合不安装插件的环境。
          </p>
          {mcpUrl ? (
            <ClientConfigPanel mcpUrl={mcpUrl} apiKey={apiKey} />
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
