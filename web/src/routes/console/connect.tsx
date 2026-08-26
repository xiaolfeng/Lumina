import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { ChannelBand, ChannelKicker } from '#/components/mcp/channel-band'
import { ClientConfigPanel } from '#/components/mcp/client-config-panel'
import { EndpointCard } from '#/components/mcp/endpoint-card'
import { ToolCatalog, WorkflowList } from '#/components/mcp/tool-catalog'
import { useMcpEndpoint } from '#/hooks/useMcpEndpoint'

export const Route = createFileRoute('/console/connect')({
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
        <ChannelBand index={2} alt>
          <ChannelKicker>客户端</ChannelKicker>
          <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
            一键复制
          </h2>
          <p className="mt-1 mb-4 text-[13px] leading-relaxed text-sea-ink-soft">
            字段名各不相同。默认
            Cursor；点名称切换片段。配置格式以客户端当前文档为准。
          </p>
          {mcpUrl ? (
            <ClientConfigPanel mcpUrl={mcpUrl} apiKey={apiKey} />
          ) : (
            <p className="text-sm text-sea-ink-soft">正在解析接入地址…</p>
          )}
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={3}>
          <ToolCatalog />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={4} alt>
          <WorkflowList />
        </ChannelBand>
      </motion.div>
    </motion.div>
  )
}
