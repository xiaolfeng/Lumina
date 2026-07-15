import { motion } from 'motion/react'
import {
  Server,
  Globe,
  Radio,
  Database,
  HardDrive,
  Cpu,
} from 'lucide-react'

import {
  fadeUp,
  sectionStagger,
  viewportOnce,
} from '@lumina/components/motion'
import { kickerBase } from './hero-section'

const shellBase =
  'border border-line bg-surface shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-sm'

export function TechSection() {
  return (
    <section className="page-wrap px-4 py-20" aria-label="架构设计">
      <motion.div
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={sectionStagger}
      >
        <div className="mb-12 text-center">
          <motion.div
            className="mb-3 flex items-center justify-center gap-3"
            variants={fadeUp}
          >
            <span className="h-px w-8 bg-gradient-to-r from-transparent to-lagoon/30" />
            <span className={kickerBase}>架构设计</span>
            <span className="h-px w-8 bg-gradient-to-l from-transparent to-lagoon/30" />
          </motion.div>
          <motion.h2
            className="display-title text-3xl font-bold text-sea-ink sm:text-4xl"
            variants={fadeUp}
          >
            三模块独立 + 统一基础设施
          </motion.h2>
        </div>

        <motion.div
          className={`${shellBase} mx-auto max-w-2xl rounded-2xl p-8`}
          variants={fadeUp}
        >
          <div className="mb-8 flex flex-wrap items-center justify-center gap-3">
            {[
              { label: 'MCP Server', icon: Server },
              { label: 'REST API', icon: Globe },
              { label: 'SSE', icon: Radio },
            ].map((item) => (
              <span
                key={item.label}
                className="inline-flex items-center gap-1.5 rounded-full bg-surface-strong px-4 py-2 text-sm font-semibold text-sea-ink shadow-sm ring-1 ring-line backdrop-blur-md transition-all duration-200 hover:shadow-md hover:ring-lagoon/20 hover:-translate-y-0.5"
              >
                <item.icon
                  className="h-4 w-4 text-lagoon"
                  aria-hidden
                />
                {item.label}
              </span>
            ))}
          </div>

          <div className="mx-auto mb-8 h-px w-3/4 bg-gradient-to-b from-line to-transparent" />

          <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
            {['RepoWiki', 'Memory', 'Q&A'].map((mod) => (
              <div
                key={mod}
                className="rounded-lg border border-lagoon/20 bg-lagoon/5 px-4 py-3 text-center text-sm font-semibold text-sea-ink"
              >
                {mod}
              </div>
            ))}
          </div>

          <div className="mx-auto mb-8 h-px w-3/4 bg-gradient-to-b from-line to-transparent" />

          <div className="flex flex-wrap items-center justify-center gap-3">
            {[
              { label: 'PostgreSQL', icon: Database },
              { label: 'Redis', icon: HardDrive },
              { label: 'LLM Provider', icon: Cpu },
            ].map((item) => (
              <span
                key={item.label}
                className="inline-flex items-center gap-1.5 rounded-full bg-surface px-4 py-2 text-sm font-semibold text-sea-ink-soft shadow-sm ring-1 ring-line backdrop-blur-md transition-all duration-200 hover:shadow-md hover:ring-palm/20 hover:-translate-y-0.5"
              >
                <item.icon
                  className="h-4 w-4 text-palm"
                  aria-hidden
                />
                {item.label}
              </span>
            ))}
          </div>
        </motion.div>
      </motion.div>
    </section>
  )
}
