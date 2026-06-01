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

export const Route = createFileRoute('/_public/')({ component: Home })

/* ─── Animation presets ────────────────────────────────── */

const fadeUp = {
  hidden: { opacity: 0, y: 18 },
  visible: { opacity: 1, y: 0 },
}

const fadeIn = {
  hidden: { opacity: 0 },
  visible: { opacity: 1 },
}

const scaleIn = {
  hidden: { opacity: 0, scale: 0.92 },
  visible: { opacity: 1, scale: 1 },
}

const heroStagger = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.12, delayChildren: 0.08 },
  },
}

const sectionStagger = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.1 },
  },
}

const viewportOnce = { once: true, margin: '-80px' } as const

/* ─── Shared class strings ─────────────────────────────── */

const shellBase =
  'border border-[var(--line)] bg-[var(--surface)] shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-sm'

const kickerBase =
  'inline-flex items-center rounded-full border border-[var(--chip-line)] px-[0.7em] py-[0.25em] text-xs font-bold uppercase tracking-widest text-[var(--lagoon-deep)]'

const kickerStyle = {
  background: 'linear-gradient(to right, transparent, var(--chip-bg), transparent)',
} as const

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
    <>
      {/* ════════ 1. HERO SECTION ════════ */}
      <motion.section
        className="relative flex min-h-[90vh] items-center justify-center overflow-hidden px-4 pt-24 pb-20 text-center md:min-h-[92vh] md:pt-28 md:pb-24"
        aria-label="主标题区域"
        initial="hidden"
        animate="visible"
        variants={heroStagger}
      >
        {/* 背景层：微妙网格 */}
        <div
          className="pointer-events-none absolute inset-0 opacity-[0.035]"
          style={{
            backgroundImage:
              'linear-gradient(rgba(42,36,32,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(42,36,32,0.4) 1px, transparent 1px)',
            backgroundSize: '48px 48px',
            maskImage: 'radial-gradient(circle at 50% 40%, black, transparent 75%)',
          }}
        />

        {/* 主光晕 */}
        <div
          className="pointer-events-none absolute top-1/2 left-1/2 h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full opacity-50 blur-3xl md:h-[900px] md:w-[900px] md:opacity-40"
          style={{ background: 'radial-gradient(circle, var(--hero-a), transparent 60%)' }}
        />

        {/* 右上余温光晕 */}
        <div
          className="pointer-events-none absolute top-[15%] right-[-5%] h-[350px] w-[350px] rounded-full opacity-30 blur-3xl md:h-[500px] md:w-[500px] md:opacity-25"
          style={{ background: 'radial-gradient(circle, var(--hero-b), transparent 55%)' }}
        />

        {/* 左下微光 */}
        <div
          className="pointer-events-none absolute bottom-[10%] left-[-5%] h-[280px] w-[280px] rounded-full opacity-20 blur-3xl md:h-[400px] md:w-[400px]"
          style={{ background: 'radial-gradient(circle, rgba(201,136,58,0.25), transparent 60%)' }}
        />

        <div className="relative z-10 mx-auto max-w-3xl">
          <motion.h1
            className="display-title mb-6 text-[clamp(4rem,10vw+1rem,7rem)] font-bold leading-[1.02] tracking-tight md:mb-8"
            variants={fadeUp}
          >
            <span className="bg-gradient-to-br from-[var(--sea-ink)] via-[var(--lagoon-deep)] to-[var(--palm)] bg-clip-text text-transparent">
              Lumina
            </span>
          </motion.h1>

          <motion.p
            className="mb-5 text-xl font-semibold text-[var(--lagoon-deep)] md:mb-6 md:text-[1.65rem]"
            variants={fadeUp}
          >
            烛照幽微，知常曰明
          </motion.p>

          <motion.p
            className="mx-auto mb-14 max-w-lg text-base leading-relaxed text-[var(--sea-ink-soft)] md:mb-16 md:max-w-xl md:text-lg"
            variants={fadeUp}
          >
            万象隐于幽微，常理没于流转。微明不求如烈日灼目，只愿燃一寸静烛：为乱麻梳理脉络，使隐晦昭然；将瞬息沉淀为常识，让过往不再是流沙。
          </motion.p>

          <motion.div
            className="flex flex-col items-center justify-center gap-4 sm:flex-row"
            variants={fadeUp}
          >
            <Button
              asChild
              size="lg"
              className="group relative h-12 overflow-hidden rounded-full bg-gradient-to-r from-[var(--lagoon)] to-[var(--palm)] px-8 text-base font-semibold !text-white shadow-lg shadow-[var(--lagoon)]/25 transition-shadow duration-300 hover:shadow-xl hover:shadow-[var(--lagoon)]/40"
            >
              <Link to="/login" aria-label="开始使用 Lumina">
                开始使用
                <ArrowRight className="ml-2 h-4 w-4 transition-transform duration-300 group-hover:translate-x-1" aria-hidden />
              </Link>
            </Button>

            <Button
              asChild
              variant="outline"
              size="lg"
              className="h-12 rounded-full border-[var(--line)] px-8 text-base font-medium text-[var(--sea-ink-soft)] transition-colors duration-300 hover:border-[var(--lagoon)]/30 hover:bg-[var(--lagoon)]/5 hover:text-[var(--sea-ink)]"
            >
              <a href="#features" aria-label="了解更多关于 Lumina 的功能">
                了解更多
              </a>
            </Button>
          </motion.div>

          {/* 底部装饰线 */}
          <motion.div
            className="mx-auto mt-20 h-px w-28 bg-gradient-to-r from-transparent via-[var(--lagoon)]/40 to-transparent md:mt-24"
            variants={fadeIn}
          />
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
            <motion.p className={`${kickerBase} mb-3`} style={kickerStyle} variants={fadeUp}>
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
                className={`${shellBase} cursor-pointer rounded-xl p-6`}
                aria-label={`${feat.title} 功能卡片`}
                variants={fadeUp}
                whileHover={{ y: -4, boxShadow: '0 8px 36px rgba(42,36,32,0.10)', transition: { duration: 0.2 } }}
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
            <motion.p className={`${kickerBase} mb-3`} style={kickerStyle} variants={fadeUp}>
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
          <motion.p className={`${kickerBase} mb-3`} style={kickerStyle} variants={fadeUp}>
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
            className={`${shellBase} overflow-hidden rounded-xl text-left`}
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
          className={`${shellBase} mx-auto max-w-xl rounded-2xl px-8 py-14`}
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
            >
              <Link to="/login" aria-label="立即开始使用 Lumina">
                立即开始
                <ArrowRight className="ml-2 h-4 w-4" aria-hidden />
              </Link>
            </Button>
          </motion.div>
        </motion.div>
      </section>
    </>
  )
}
