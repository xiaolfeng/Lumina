import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
  BookOpen,
  Brain,
  CheckCircle2,
  MessageCircle,
  Terminal,
  ArrowRight,
  Sparkles,
  Settings,
  Plug,
} from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Link } from '@tanstack/react-router'

export const Route = createFileRoute('/_public/start')({
  component: StartPage,
})

/* ─── Animation presets ────────────────────────────────── */

const fadeUp = {
  hidden: { opacity: 0, y: 18 },
  visible: { opacity: 1, y: 0 },
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
  'text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--lagoon-deep)]'

/* ─── Data ─────────────────────────────────────────────── */

const steps = [
  {
    step: 1,
    icon: Settings,
    title: '配置环境',
    description: '复制环境变量模板并填写必要的配置信息。',
    code: 'cp .env.example .env',
    detail:
      '编辑 .env 文件，配置 PostgreSQL、Redis 连接信息以及 LLM Provider 参数。开发环境下大多数选项已有合理默认值。',
  },
  {
    step: 2,
    icon: Terminal,
    title: '安装依赖',
    description: '安装 Go 后端依赖。',
    code: 'go mod tidy',
    detail: '确保本地已安装 Go 1.25+，运行 mod tidy 拉取所有依赖。',
  },
  {
    step: 3,
    icon: Sparkles,
    title: '启动服务',
    description: '生成 Swagger 文档并启动后端服务。',
    code: 'make dev',
    detail:
      '推荐使用 make dev 命令，它会自动生成 API 文档并启动服务。服务默认监听 0.0.0.0:8080。',
  },
  {
    step: 4,
    icon: Plug,
    title: '接入 MCP',
    description: '在你的 AI Agent 中配置 Lumina MCP Server。',
    code: 'lumina-mcp --transport streamable --port 8080',
    detail:
      '通过 Streamable MCP 协议，任何支持 MCP 的 AI Agent 均可接入。支持 RepoWiki 分析、Memory 记忆、Q&A 问答三大工具集。',
  },
] as const

const coreModules = [
  {
    icon: BookOpen,
    title: 'RepoWiki',
    description: '克隆项目并通过 LLM 分析生成结构化 Wiki 文档。',
    tools: ['repoWiki_analyze', 'repoWiki_list'],
  },
  {
    icon: Brain,
    title: 'Memory',
    description: 'AI 的长期决策记忆，跨会话保留重要约定与决策。',
    tools: ['memory_create', 'memory_search', 'memory_list'],
  },
  {
    icon: MessageCircle,
    title: 'Q&A',
    description: 'Agent 与用户的富交互式问答通道，SSE 实时推送。',
    tools: ['qa_pushQuestion', 'qa_collectAnswer'],
  },
] as const

/* ─── Component ────────────────────────────────────────── */

function StartPage() {
  return (
    <>
      {/* ════════ HERO ════════ */}
      <section
        className="page-wrap px-4 pb-12 pt-12 md:pb-16 md:pt-16"
        aria-label="开始使用"
      >
        <motion.div
          className="mx-auto max-w-3xl text-center"
          initial="hidden"
          animate="visible"
          variants={sectionStagger}
        >
          <motion.div
            className="mb-4 flex items-center justify-center gap-3"
            variants={fadeUp}
          >
            <span className="h-px w-8 bg-gradient-to-r from-transparent to-[var(--lagoon)]/30" />
            <span className={kickerBase}>快速开始</span>
            <span className="h-px w-8 bg-gradient-to-l from-transparent to-[var(--lagoon)]/30" />
          </motion.div>

          <motion.h1
            className="display-title mb-4 text-4xl font-bold text-[var(--sea-ink)] sm:text-5xl"
            variants={fadeUp}
          >
            开始使用{' '}
            <span className="bg-gradient-to-br from-[var(--lagoon-deep)] to-[var(--palm)] bg-clip-text text-transparent">
              Lumina
            </span>
          </motion.h1>

          <motion.p
            className="mx-auto max-w-lg text-base leading-relaxed text-[var(--sea-ink-soft)] md:text-lg"
            variants={fadeUp}
          >
            几步即可完成部署，让 AI 获得深度代码认知与长期记忆。
          </motion.p>
        </motion.div>
      </section>

      {/* ════════ STEPS ════════ */}
      <section className="page-wrap px-4 pb-20" aria-label="部署步骤">
        <motion.div
          className="mx-auto max-w-2xl space-y-6"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          {steps.map((item) => (
            <motion.article
              key={item.step}
              className={`${shellBase} rounded-xl p-6`}
              aria-label={`步骤 ${item.step}：${item.title}`}
              variants={fadeUp}
              whileHover={{
                boxShadow: '0 8px 36px rgba(42,36,32,0.10)',
                transition: { duration: 0.2 },
              }}
            >
              <div className="mb-4 flex items-start gap-4">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[var(--lagoon)]/10">
                  <span className="display-title text-sm font-bold text-[var(--lagoon-deep)]">
                    {item.step}
                  </span>
                </div>
                <div className="flex-1">
                  <div className="mb-1 flex items-center gap-2">
                    <item.icon
                      className="h-4 w-4 text-[var(--lagoon)]"
                      aria-hidden
                    />
                    <h3 className="text-lg font-semibold text-[var(--sea-ink)]">
                      {item.title}
                    </h3>
                  </div>
                  <p className="text-sm text-[var(--sea-ink-soft)]">
                    {item.description}
                  </p>
                </div>
              </div>

              <div className="overflow-hidden rounded-lg border border-[var(--line)] bg-[var(--foam)]">
                <div className="flex items-center gap-2 border-b border-[var(--line)] px-4 py-2">
                  <Terminal
                    className="h-3.5 w-3.5 text-[var(--sea-ink-soft)]"
                    aria-hidden
                  />
                  <span className="text-[11px] font-semibold uppercase tracking-wider text-[var(--sea-ink-soft)]">
                    终端
                  </span>
                </div>
                <pre className="overflow-x-auto px-4 py-3">
                  <code className="text-sm text-[var(--sea-ink)]">
                    <span className="text-[var(--lagoon-deep)]">$</span>{' '}
                    {item.code}
                  </code>
                </pre>
              </div>

              <p className="mt-3 text-sm leading-relaxed text-[var(--sea-ink-soft)]">
                {item.detail}
              </p>
            </motion.article>
          ))}
        </motion.div>
      </section>

      {/* ════════ VERIFY ════════ */}
      <section className="page-wrap px-4 pb-20" aria-label="验证部署">
        <motion.div
          className="mx-auto max-w-2xl"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <motion.div
            className="mb-8 flex items-center justify-center gap-3"
            variants={fadeUp}
          >
            <span className="h-px w-8 bg-gradient-to-r from-transparent to-[var(--lagoon)]/30" />
            <span className={kickerBase}>验证部署</span>
            <span className="h-px w-8 bg-gradient-to-l from-transparent to-[var(--lagoon)]/30" />
          </motion.div>

          <motion.div
            className={`${shellBase} rounded-xl p-6`}
            variants={fadeUp}
          >
            <p className="mb-4 text-sm text-[var(--sea-ink-soft)]">
              服务启动后，使用以下命令验证部署是否成功：
            </p>
            <div className="overflow-hidden rounded-lg border border-[var(--line)] bg-[var(--foam)]">
              <div className="flex items-center gap-2 border-b border-[var(--line)] px-4 py-2">
                <Terminal
                  className="h-3.5 w-3.5 text-[var(--sea-ink-soft)]"
                  aria-hidden
                />
                <span className="text-[11px] font-semibold uppercase tracking-wider text-[var(--sea-ink-soft)]">
                  终端
                </span>
              </div>
              <pre className="overflow-x-auto px-4 py-3">
                <code className="text-sm text-[var(--sea-ink)]">
                  <span className="text-[var(--lagoon-deep)]">$</span> curl
                  http://localhost:8080/api/v1/health/ping
                  {'\n'}
                  <span className="text-[var(--lagoon-deep)]">{'>'}</span>{' '}
                  {'{'}"status":"ok"{'}'}
                </code>
              </pre>
            </div>

            <div className="mt-4 flex items-center gap-2 rounded-lg bg-green-50 px-3 py-2 dark:bg-green-900/10">
              <CheckCircle2
                className="h-4 w-4 shrink-0 text-green-600 dark:text-green-400"
                aria-hidden
              />
              <span className="text-sm font-medium text-green-700 dark:text-green-300">
                收到 ok 响应即表示部署成功
              </span>
            </div>
          </motion.div>
        </motion.div>
      </section>

      {/* ════════ CORE MODULES ════════ */}
      <section className="page-wrap px-4 pb-20" aria-label="核心模块">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <motion.div
            className="mb-8 flex items-center justify-center gap-3"
            variants={fadeUp}
          >
            <span className="h-px w-8 bg-gradient-to-r from-transparent to-[var(--lagoon)]/30" />
            <span className={kickerBase}>核心模块</span>
            <span className="h-px w-8 bg-gradient-to-l from-transparent to-[var(--lagoon)]/30" />
          </motion.div>

          <motion.h2
            className="display-title mb-8 text-center text-2xl font-bold text-[var(--sea-ink)] sm:text-3xl"
            variants={fadeUp}
          >
            三大能力，统一接入
          </motion.h2>

          <div className="mx-auto grid max-w-3xl grid-cols-1 gap-6 sm:grid-cols-3">
            {coreModules.map((mod) => (
              <motion.article
                key={mod.title}
                className={`${shellBase} cursor-pointer rounded-xl p-5`}
                aria-label={`${mod.title} 模块说明`}
                variants={fadeUp}
                whileHover={{
                  y: -4,
                  boxShadow: '0 8px 36px rgba(42,36,32,0.10)',
                  transition: { duration: 0.2 },
                }}
              >
                <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-[var(--lagoon)]/10">
                  <mod.icon
                    className="h-5 w-5 text-[var(--lagoon-deep)]"
                    aria-hidden
                  />
                </div>
                <h3 className="mb-2 text-base font-semibold text-[var(--sea-ink)]">
                  {mod.title}
                </h3>
                <p className="mb-3 text-sm leading-relaxed text-[var(--sea-ink-soft)]">
                  {mod.description}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {mod.tools.map((tool) => (
                    <span
                      key={tool}
                      className="inline-flex items-center rounded-full bg-gradient-to-r from-[var(--lagoon)]/5 to-[var(--palm)]/5 px-2.5 py-0.5 text-[10px] font-semibold text-[var(--lagoon-deep)] ring-1 ring-[var(--lagoon)]/10"
                    >
                      {tool}
                    </span>
                  ))}
                </div>
              </motion.article>
            ))}
          </div>
        </motion.div>
      </section>

      {/* ════════ CTA ════════ */}
      <section
        className="page-wrap px-4 pb-24 text-center"
        aria-label="行动号召"
      >
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
            准备好了吗？
          </motion.h2>
          <motion.p
            className="mb-8 text-sm text-[var(--sea-ink-soft)]"
            variants={fadeUp}
          >
            登录后即可开始使用 Lumina 的全部能力。
          </motion.p>
          <motion.div
            className="flex flex-col items-center justify-center gap-3 sm:flex-row"
            variants={fadeUp}
          >
            <Button
              asChild
              size="lg"
              className="group relative h-12 overflow-hidden rounded-full bg-gradient-to-r from-[var(--lagoon)] to-[var(--palm)] px-8 text-base font-semibold !text-white shadow-lg shadow-[var(--lagoon)]/25 transition-shadow duration-300 hover:shadow-xl hover:shadow-[var(--lagoon)]/40"
            >
              <Link to="/auth/login" aria-label="登录 Lumina">
                立即登录
                <ArrowRight
                  className="ml-2 h-4 w-4 transition-transform duration-300 group-hover:translate-x-1"
                  aria-hidden
                />
              </Link>
            </Button>
            <Button
              asChild
              variant="outline"
              size="lg"
              className="h-12 rounded-full border-[var(--line)] px-8 text-base font-medium text-[var(--sea-ink-soft)] transition-colors duration-300 hover:border-[var(--lagoon)]/30 hover:bg-[var(--lagoon)]/5 hover:text-[var(--sea-ink)]"
            >
              <a
                href="https://github.com/xiaolfeng/Lumina"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="查看 GitHub 仓库"
              >
                查看源码
              </a>
            </Button>
          </motion.div>
        </motion.div>
      </section>
    </>
  )
}
