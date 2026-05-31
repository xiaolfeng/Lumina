import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
  BookOpen,
  Brain,
  MessageCircle,
  ArrowRight,
  Sparkles,
  Server,
  Globe,
  Radio,
  Database,
  HardDrive,
  Cpu,
} from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Navbar } from '#/components/Navbar'
import { Footer } from '#/components/Footer'

export const Route = createFileRoute('/')({ component: Home })

/* ─── Animation presets ────────────────────────────────── */

const fadeUp = {
  hidden: { opacity: 0, y: 18 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: 'easeOut' },
  },
}

const fadeIn = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { duration: 0.4, ease: 'easeOut' },
  },
}

const scaleIn = {
  hidden: { opacity: 0, scale: 0.92 },
  visible: {
    opacity: 1,
    scale: 1,
    transition: { duration: 0.4, ease: 'easeOut' },
  },
}

const heroStagger = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.1 },
  },
}

const sectionStagger = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.12 },
  },
}

const viewportOnce = { once: true, margin: '-80px' } as const

/* ─── Data ─────────────────────────────────────────────── */

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

/* ─── Component ────────────────────────────────────────── */

function Home() {
  return (
    <div className="min-h-screen">
      <Navbar />

      {/* ════════ 1. HERO SECTION ════════ */}
      <motion.section
        className="page-wrap px-4 pt-32 pb-20 text-center"
        aria-label="主标题区域"
        initial="hidden"
        animate="visible"
        variants={heroStagger}
      >
        <div className="mx-auto max-w-3xl">
          <motion.p className="island-kicker mb-4" variants={fadeUp}>
            AI 知识中枢
          </motion.p>

          <motion.h1
            className="display-title mb-6 text-[clamp(2.8rem,7vw+1rem,5.2rem)] font-bold leading-[1.08] tracking-tight text-[var(--sea-ink)]"
            variants={fadeUp}
          >
            Lumina
          </motion.h1>

          <motion.p
            className="mb-4 text-xl font-semibold text-[var(--lagoon-deep)]"
            variants={fadeUp}
          >
            烛照幽微，知常曰明
          </motion.p>

          <motion.p
            className="mx-auto mb-10 max-w-2xl text-base leading-relaxed text-[var(--sea-ink-soft)]"
            variants={fadeUp}
          >
            万象隐于幽微，常理没于流转。人常在纷繁中迷失头绪，于遗忘里重复跋涉。微明不求如烈日灼目，只愿燃一寸静烛：为乱麻梳理脉络，使隐晦昭然；将瞬息沉淀为常识，让过往不再是流沙。愿这微光渡过无形的桥，照亮每一次探寻。
          </motion.p>

          <motion.div
            className="flex flex-wrap items-center justify-center gap-4"
            variants={fadeUp}
          >
            <Button
              asChild
              size="lg"
              className="touch-action-manipulation cursor-pointer"
            >
              <Link to="/login" aria-label="开始使用 Lumina">
                开始使用
                <ArrowRight className="ml-2 h-4 w-4" aria-hidden />
              </Link>
            </Button>

            <Button
              asChild
              variant="outline"
              size="lg"
              className="touch-action-manipulation cursor-pointer"
            >
              <a href="#features" aria-label="了解更多关于 Lumina 的功能">
                了解更多
              </a>
            </Button>
          </motion.div>
        </div>
      </motion.section>

      {/* ════════ 2. FEATURES SECTION ════════ */}
      <section id="features" className="page-wrap px-4 py-20" aria-label="核心功能">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <div className="mb-12 text-center">
            <motion.p className="island-kicker mb-3" variants={fadeUp}>
              核心能力
            </motion.p>
            <motion.h2
              className="display-title text-3xl font-bold text-[var(--sea-ink)] sm:text-4xl"
              variants={fadeUp}
            >
              三大独立领域，统一知识中枢
            </motion.h2>
          </div>

          <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
            {features.map((feat) => (
              <motion.article
                key={feat.title}
                className="island-shell cursor-pointer rounded-xl border p-6"
                aria-label={`${feat.title} 功能卡片`}
                variants={fadeUp}
                whileHover={{ y: -4, transition: { duration: 0.2 } }}
              >
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-[var(--lagoon)]/10">
                  <feat.icon
                    className="h-6 w-6 text-[var(--lagoon-deep)]"
                    aria-hidden
                  />
                </div>

                <h3 className="mb-2 text-lg font-semibold text-[var(--sea-ink)]">
                  {feat.title}
                </h3>

                <p className="mb-4 text-sm leading-relaxed text-[var(--sea-ink-soft)]">
                  {feat.description}
                </p>

                <div className="flex flex-wrap gap-2">
                  {feat.tags.map((tag) => (
                    <span
                      key={tag}
                      className="rounded-full border border-[var(--chip-line)] bg-[var(--chip-bg)] px-3 py-1 text-xs font-medium text-[var(--sea-ink-soft)]"
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

      {/* ════════ 3. ARCHITECTURE SECTION ════════ */}
      <section className="page-wrap px-4 py-20" aria-label="架构设计">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <div className="mb-12 text-center">
            <motion.p className="island-kicker mb-3" variants={fadeUp}>
              架构设计
            </motion.p>
            <motion.h2
              className="display-title text-3xl font-bold text-[var(--sea-ink)] sm:text-4xl"
              variants={fadeUp}
            >
              三模块独立 + 统一基础设施
            </motion.h2>
          </div>

          <motion.div
            className="island-shell mx-auto max-w-2xl rounded-2xl p-8"
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
                  className="inline-flex items-center gap-1.5 rounded-full border border-[var(--chip-line)] bg-[var(--chip-bg)] px-4 py-2 text-sm font-medium text-[var(--sea-ink)]"
                >
                  <item.icon
                    className="h-4 w-4 text-[var(--lagoon)]"
                    aria-hidden
                  />
                  {item.label}
                </span>
              ))}
            </div>

            <div className="mx-auto mb-8 h-px w-3/4 bg-gradient-to-b from-[var(--line)] to-transparent" />

            <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
              {['RepoWiki', 'Memory', 'Q&A'].map((mod) => (
                <div
                  key={mod}
                  className="rounded-lg border border-[var(--lagoon)]/20 bg-[var(--lagoon)]/5 px-4 py-3 text-center text-sm font-semibold text-[var(--sea-ink)]"
                >
                  {mod}
                </div>
              ))}
            </div>

            <div className="mx-auto mb-8 h-px w-3/4 bg-gradient-to-b from-[var(--line)] to-transparent" />

            <div className="flex flex-wrap items-center justify-center gap-3">
              {[
                { label: 'PostgreSQL', icon: Database },
                { label: 'Redis', icon: HardDrive },
                { label: 'LLM Provider', icon: Cpu },
              ].map((item) => (
                <span
                  key={item.label}
                  className="inline-flex items-center gap-1.5 rounded-full border border-[var(--chip-line)] bg-[var(--chip-bg)] px-4 py-2 text-sm font-medium text-[var(--sea-ink-soft)]"
                >
                  <item.icon
                    className="h-4 w-4 text-[var(--palm)]"
                    aria-hidden
                  />
                  {item.label}
                </span>
              ))}
            </div>
          </motion.div>
        </motion.div>
      </section>

      {/* ════════ 4. MCP SECTION ════════ */}
      <section className="page-wrap px-4 py-20" aria-label="MCP 协议">
        <motion.div
          className="mx-auto max-w-2xl text-center"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <motion.p className="island-kicker mb-3" variants={fadeUp}>
            开放协议
          </motion.p>
          <motion.h2
            className="display-title mb-4 text-3xl font-bold text-[var(--sea-ink)] sm:text-4xl"
            variants={fadeUp}
          >
            开放协议，无缝接入
          </motion.h2>
          <motion.p
            className="mb-8 text-base leading-relaxed text-[var(--sea-ink-soft)]"
            variants={fadeUp}
          >
            通过 Streamable MCP 标准协议，任何支持 MCP 的 AI Agent 均可接入
            Lumina 的全部能力。
          </motion.p>

          <motion.div
            className="island-shell overflow-hidden rounded-xl text-left"
            variants={scaleIn}
          >
            <div className="border-b border-[var(--line)] bg-[var(--foam)] px-4 py-2.5">
              <span className="text-xs font-medium text-[var(--sea-ink-soft)]">
                MCP Tools
              </span>
            </div>
            <pre className="overflow-x-auto px-5 py-4">
              <code className="text-sm leading-relaxed text-[var(--sea-ink)]">
                {mcpTools.map((tool, i) => (
                  <span key={tool}>
                    {i > 0 && '\n'}
                    <span className="text-[var(--lagoon-deep)]">{'>'}</span>{' '}
                    {tool}
                  </span>
                ))}
              </code>
            </pre>
          </motion.div>
        </motion.div>
      </section>

      {/* ════════ 5. CTA SECTION ════════ */}
      <section className="page-wrap px-4 py-24 text-center" aria-label="行动号召">
        <motion.div
          className="island-shell mx-auto max-w-xl rounded-2xl px-8 py-14"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <motion.div variants={fadeUp}>
            <Sparkles
              className="mx-auto mb-5 h-10 w-10 text-[var(--lagoon)]"
              aria-hidden
            />
          </motion.div>
          <motion.h2
            className="display-title mb-4 text-2xl font-bold text-[var(--sea-ink)] sm:text-3xl"
            variants={fadeUp}
          >
            准备好让 AI 真正理解你的代码了吗？
          </motion.h2>
          <motion.p
            className="mb-8 text-sm text-[var(--sea-ink-soft)]"
            variants={fadeUp}
          >
            立即开始使用 Lumina，开启 AI 深度代码认知之旅。
          </motion.p>
          <motion.div variants={fadeUp}>
            <Button
              asChild
              size="lg"
              className="touch-action-manipulation cursor-pointer"
            >
              <Link to="/login" aria-label="立即开始使用 Lumina">
                立即开始
                <ArrowRight className="ml-2 h-4 w-4" aria-hidden />
              </Link>
            </Button>
          </motion.div>
        </motion.div>
      </section>

      {/* ════════ 6. FOOTER ════════ */}
      <Footer />
    </div>
  )
}
