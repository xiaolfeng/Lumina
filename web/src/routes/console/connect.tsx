import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { ChannelBand } from '#/components/mcp/channel-band'
import { ManualConnectPanel } from '#/components/mcp/manual-connect-panel'
import { PluginInstallPanel } from '#/components/mcp/plugin-install-panel'
import { SkillsInstallPanel } from '#/components/mcp/skills-install-panel'
import { ToolCatalog, WorkflowList } from '#/components/mcp/tool-catalog'
import { useMcpEndpoint } from '#/hooks/useMcpEndpoint'

export const Route = createFileRoute('/console/connect')({
  staticData: { crumb: '接入指南' },
  component: ConnectPage,
})

function ConnectPage() {
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
          <PluginInstallPanel origin={origin} />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={2} alt>
          <ManualConnectPanel mcpUrl={mcpUrl} origin={origin} source={source} />
        </ChannelBand>
      </motion.div>

      <motion.div variants={staggerItem}>
        <ChannelBand index={3}>
          <SkillsInstallPanel origin={origin} />
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
