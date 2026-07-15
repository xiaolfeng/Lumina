import { motion } from 'motion/react'
import { BookOpen, Brain, MessageCircle } from 'lucide-react'

import {
  fadeUp,
  scaleIn,
  sectionStagger,
  viewportOnce,
} from '@lumina/components/motion'
import { kickerBase } from './hero-section'

const shellBase =
  'border border-line bg-surface shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-sm'

const features = [
  {
    icon: BookOpen,
    title: 'RepoWiki',
    description:
      '克隆项目并通过 LLM 分析生成结构化 Wiki 文档，让 AI 深度理解代码库结构、约定与设计意图。',
    tags: ['Git 克隆', 'LLM 分析', 'Wiki 生成'],
  },
  {
    icon: Brain,
    title: 'Memory',
    description:
      'AI 的长期决策记忆，跨会话保留重要约定与决策，越用越懂你的编码风格与项目约定。',
    tags: ['决策卡片', '标签分类', '条件检索'],
  },
  {
    icon: MessageCircle,
    title: 'Q&A',
    description:
      'Agent 与用户的富交互式问答通道，支持选项、文本、分批推送，SSE 实时响应。',
    tags: ['Session 管理', 'SSE 推送', '富交互'],
  },
] as const

const mcpTools = ['repoWiki_analyze', 'memory_create', 'qa_pushQuestion']

export function FeaturesSection() {
  return (
    <section
      id="features"
      className="page-wrap px-4 py-20"
      aria-label="核心功能"
    >
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
            <span className={kickerBase}>核心能力</span>
            <span className="h-px w-8 bg-gradient-to-l from-transparent to-lagoon/30" />
          </motion.div>
          <motion.h2
            className="display-title text-3xl font-bold text-sea-ink sm:text-4xl"
            variants={fadeUp}
          >
            三大独立领域，统一知识中枢
          </motion.h2>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {features.map((feat) => (
            <motion.article
              key={feat.title}
              className={`${shellBase} cursor-pointer rounded-xl p-6`}
              aria-label={`${feat.title} 功能卡片`}
              variants={fadeUp}
              whileHover={{
                y: -4,
                boxShadow: '0 8px 36px rgba(42,36,32,0.10)',
                transition: { duration: 0.2 },
              }}
            >
              <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-lagoon/10">
                <feat.icon
                  className="h-6 w-6 text-lagoon-deep"
                  aria-hidden
                />
              </div>

              <h3 className="mb-2 text-lg font-semibold text-sea-ink">
                {feat.title}
              </h3>

              <p className="mb-4 text-sm leading-relaxed text-sea-ink-soft">
                {feat.description}
              </p>

              <div className="flex flex-wrap gap-2">
                {feat.tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-gradient-to-r from-lagoon/5 to-palm/5 px-3 py-1 text-[11px] font-semibold text-lagoon-deep ring-1 ring-lagoon/10 backdrop-blur-sm transition-all duration-200 hover:from-lagoon/10 hover:to-palm/10 hover:ring-lagoon/25"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </motion.article>
          ))}
        </div>
      </motion.div>
    </section>
  )
}

export function McpSection() {
  return (
    <section className="page-wrap px-4 py-20" aria-label="MCP 协议">
      <motion.div
        className="mx-auto max-w-2xl text-center"
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={sectionStagger}
      >
        <motion.div
          className="mb-3 flex items-center justify-center gap-3"
          variants={fadeUp}
        >
          <span className="h-px w-8 bg-gradient-to-r from-transparent to-lagoon/30" />
          <span className={kickerBase}>开放协议</span>
          <span className="h-px w-8 bg-gradient-to-l from-transparent to-lagoon/30" />
        </motion.div>
        <motion.h2
          className="display-title mb-4 text-3xl font-bold text-sea-ink sm:text-4xl"
          variants={fadeUp}
        >
          开放协议，无缝接入
        </motion.h2>
        <motion.p
          className="mb-8 text-base leading-relaxed text-sea-ink-soft"
          variants={fadeUp}
        >
          通过 Streamable MCP 标准协议，任何支持 MCP 的 AI Agent 均可接入
          Lumina 的全部能力。
        </motion.p>

        <motion.div
          className={`${shellBase} overflow-hidden rounded-xl text-left`}
          variants={scaleIn}
        >
          <div className="flex items-center gap-2 border-b border-line bg-foam px-4 py-2.5">
            <span className="inline-block h-2 w-2 rounded-full bg-lagoon/60" />
            <span className="text-[11px] font-semibold uppercase tracking-wider text-sea-ink-soft">
              MCP Tools
            </span>
          </div>
          <pre className="overflow-x-auto px-5 py-4">
            <code className="text-sm leading-relaxed text-sea-ink">
              {mcpTools.map((tool, i) => (
                <span key={tool}>
                  {i > 0 && '\n'}
                  <span className="text-lagoon-deep">
                    {'>'}
                  </span>{' '}
                  {tool}
                </span>
              ))}
            </code>
          </pre>
        </motion.div>
      </motion.div>
    </section>
  )
}
