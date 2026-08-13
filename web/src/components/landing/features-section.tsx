import { motion } from 'motion/react'
import { GitBranch, Brain, MessageCircle, Pin } from 'lucide-react'

import {
  fadeUp,
  sectionStagger,
  viewportOnce,
} from '@lumina/components/motion'

const modules = [
  {
    icon: GitBranch,
    name: 'RepoWiki',
    description: '克隆项目，5 角色协作生成结构化 Wiki 文档。',
    tag: '已实现',
  },
  {
    icon: Brain,
    name: 'Memory',
    description: 'AI 的长期决策记忆，MCP 端主动推送构建。',
    tag: '设计中',
  },
  {
    icon: MessageCircle,
    name: 'Q&A',
    description: 'Agent 与用户的富交互式问答，WebSocket 实时推送。',
    tag: '已实现',
  },
  {
    icon: Pin,
    name: 'Pin',
    description: '跨项目依赖约束传递，FIFO 队列定向消费。',
    tag: '已实现',
  },
] as const

export function FeaturesSection() {
  return (
    <section id="modules" className="page-wrap px-6 py-16" aria-label="核心模块">
      <motion.div
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={sectionStagger}
        className="grid grid-cols-1 gap-px border border-line bg-line sm:grid-cols-2 lg:grid-cols-4"
      >
        {modules.map((mod) => (
          <motion.article
            key={mod.name}
            variants={fadeUp}
            className="flex flex-col bg-background p-8"
            aria-label={`${mod.name} 模块`}
          >
            <div className="flex h-9 w-9 items-center justify-center border border-line text-lagoon-deep">
              <mod.icon className="h-5 w-5" aria-hidden />
            </div>

            <h3 className="display-title mt-6 text-lg font-semibold text-sea-ink">
              {mod.name}
            </h3>

            <p className="mt-3 text-sm leading-relaxed text-sea-ink-soft">
              {mod.description}
            </p>

            <p className="mt-5 text-[10px] font-bold uppercase tracking-[0.12em] text-sea-ink-soft">
              {mod.tag}
            </p>
          </motion.article>
        ))}
      </motion.div>
    </section>
  )
}
