import { motion } from 'motion/react'

import {
  fadeUp,
  sectionStagger,
  viewportOnce,
} from '@lumina/components/motion'

const channels = [
  {
    kicker: '对外通道',
    name: 'MCP Server',
    description: 'Streamable MCP 协议，20+ 工具供 Agent 编排调用。',
  },
  {
    kicker: '对外通道',
    name: 'REST API',
    description: 'HTTP 接口，供控制台前端与 Wiki Reader 消费。',
  },
  {
    kicker: '对外通道',
    name: 'WebSocket',
    description: 'Q&A 实时问题推送，心跳检测与会话恢复。',
  },
] as const

export function TechSection() {
  return (
    <section className="page-wrap px-6 py-16" aria-label="对外通道">
      <motion.div
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={sectionStagger}
        className="grid grid-cols-1 gap-px border border-line bg-line sm:grid-cols-3"
      >
        {channels.map((ch) => (
          <motion.div
            key={ch.name}
            variants={fadeUp}
            className="bg-background p-8"
          >
            <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-lagoon-deep">
              {ch.kicker}
            </p>
            <h3 className="display-title mt-3 text-xl font-semibold text-sea-ink">
              {ch.name}
            </h3>
            <p className="mt-3 text-sm leading-relaxed text-sea-ink-soft">
              {ch.description}
            </p>
          </motion.div>
        ))}
      </motion.div>
    </section>
  )
}
