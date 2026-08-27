import { useEffect, useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
  ArrowRight,
  BookOpen,
  Brain,
  CheckCircle2,
  FolderKanban,
  MessageCircle,
  MonitorPlay,
  Pin,
  Plug,
  Settings,
  Sparkles,
  Terminal,
} from 'lucide-react'

import { Button } from '@lumina/components/ui/button'
import { CopyBlock } from '#/components/mcp/copy-block'
import {
  MCP_PATH,
  MCP_TOOL_MODULES,
  buildAuthorizationHeader,
  buildClientSnippet,
  buildMcpUrl,
} from '#/lib/mcp-connect'
import { buildClaudePluginSnippet } from '#/lib/plugin-connect'

export const Route = createFileRoute('/_public/start')({
  component: StartPage,
})

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

const shellBase = 'border border-line bg-surface'

const kickerBase =
  'text-[11px] font-semibold uppercase tracking-[0.2em] text-lagoon-deep'

const deploySteps = [
  {
    step: 1,
    icon: Settings,
    title: '配置环境',
    description: '复制环境变量模板并填写必要的配置信息。',
    code: 'cp .env.example .env',
    detail:
      '编辑 .env 文件，配置 PostgreSQL、Redis 以及 LLM_ENCRYPT_SECRET。开发环境下大多数选项已有合理默认值。',
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
      'make dev 会生成 API 文档并启动当前进程。MCP 端点随 HTTP 服务一起暴露，不存在独立的 lumina-mcp 命令。默认监听 0.0.0.0:8080。',
  },
] as const

function iconForModule(id: string) {
  switch (id) {
    case 'project':
      return FolderKanban
    case 'qa':
      return MessageCircle
    case 'preview':
      return MonitorPlay
    case 'pin':
      return Pin
    default:
      return BookOpen
  }
}

function StartPage() {
  const [mcpUrl, setMcpUrl] = useState(`https://<host>${MCP_PATH}`)

  useEffect(() => {
    setMcpUrl(buildMcpUrl(window.location.origin))
  }, [])

  const cursorSnippet = buildClientSnippet('cursor', { mcpUrl })
  const pluginSnippet = buildClaudePluginSnippet(
    mcpUrl.slice(0, -MCP_PATH.length),
  )

  return (
    <>
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
            <span className="h-px w-8 bg-line" />
            <span className={kickerBase}>快速开始</span>
            <span className="h-px w-8 bg-line" />
          </motion.div>

          <motion.h1
            className="display-title mb-4 text-4xl font-bold text-sea-ink sm:text-5xl"
            variants={fadeUp}
          >
            开始使用 <span className="text-lagoon-deep">Lumina</span>
          </motion.h1>

          <motion.p
            className="mx-auto max-w-lg text-base leading-relaxed text-sea-ink-soft md:text-lg"
            variants={fadeUp}
          >
            服务启动后，选择最适合你的接入方式。Claude Code
            可以直接安装插件，其他客户端也能通过 MCP 连接当前 Lumina。
          </motion.p>
        </motion.div>
      </section>

      <section className="page-wrap px-4 pb-20" aria-label="部署步骤">
        <motion.div
          className="mx-auto max-w-2xl space-y-6"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          {deploySteps.map((item) => (
            <motion.article
              key={item.step}
              className={`${shellBase} p-6`}
              aria-label={`步骤 ${item.step}：${item.title}`}
              variants={fadeUp}
              whileHover={{
                boxShadow: '0 8px 36px rgba(51,39,28,0.10)',
                transition: { duration: 0.2 },
              }}
            >
              <div className="mb-4 flex items-start gap-4">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center bg-lagoon/10">
                  <span className="display-title text-sm font-bold text-lagoon-deep">
                    {item.step}
                  </span>
                </div>
                <div className="flex-1">
                  <div className="mb-1 flex items-center gap-2">
                    <item.icon className="h-4 w-4 text-lagoon" aria-hidden />
                    <h3 className="text-lg font-semibold text-sea-ink">
                      {item.title}
                    </h3>
                  </div>
                  <p className="text-sm text-sea-ink-soft">
                    {item.description}
                  </p>
                </div>
              </div>

              <div className="overflow-hidden border border-line bg-foam">
                <div className="flex items-center gap-2 border-b border-line px-4 py-2">
                  <Terminal
                    className="h-3.5 w-3.5 text-sea-ink-soft"
                    aria-hidden
                  />
                  <span className="text-[11px] font-semibold uppercase tracking-wider text-sea-ink-soft">
                    终端
                  </span>
                </div>
                <pre className="overflow-x-auto px-4 py-3">
                  <code className="text-sm text-sea-ink">
                    <span className="text-lagoon-deep">$</span> {item.code}
                  </code>
                </pre>
              </div>

              <p className="mt-3 text-sm leading-relaxed text-sea-ink-soft">
                {item.detail}
              </p>
            </motion.article>
          ))}
        </motion.div>
      </section>

      <section className="page-wrap px-4 pb-20" aria-label="验证部署">
        <motion.div
          className="mx-auto max-w-2xl"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <SectionKicker>验证部署</SectionKicker>
          <motion.div className={`${shellBase} p-6`} variants={fadeUp}>
            <p className="mb-4 text-sm text-sea-ink-soft">
              服务启动后，使用以下命令验证部署是否成功：
            </p>
            <div className="overflow-hidden border border-line bg-foam">
              <div className="flex items-center gap-2 border-b border-line px-4 py-2">
                <Terminal
                  className="h-3.5 w-3.5 text-sea-ink-soft"
                  aria-hidden
                />
                <span className="text-[11px] font-semibold uppercase tracking-wider text-sea-ink-soft">
                  终端
                </span>
              </div>
              <pre className="overflow-x-auto px-4 py-3">
                <code className="text-sm text-sea-ink">
                  <span className="text-lagoon-deep">$</span> curl
                  http://localhost:8080/api/v1/health/ping
                  {'\n'}
                  <span className="text-lagoon-deep">{'>'}</span> {'{'}
                  "status":"ok"{'}'}
                </code>
              </pre>
            </div>
            <div className="mt-4 flex items-center gap-2 bg-green-50 px-3 py-2 dark:bg-green-900/10">
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

      <section className="page-wrap px-4 pb-20" aria-label="接入 Lumina">
        <motion.div
          className="mx-auto max-w-4xl"
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <SectionKicker>连接你的 AI 助手</SectionKicker>
          <motion.h2
            className="display-title mb-3 text-center text-2xl font-bold text-sea-ink sm:text-3xl"
            variants={fadeUp}
          >
            插件优先，独立 MCP 也一直保留
          </motion.h2>
          <motion.p
            className="mx-auto mb-7 max-w-2xl text-center text-sm leading-relaxed text-sea-ink-soft"
            variants={fadeUp}
          >
            两种方式连接的是同一个 Lumina。Claude Code
            用户安装插件最省事；Cursor、Windsurf、VS Code、Codex、Grok
            等客户端可以直接配置 MCP。
          </motion.p>

          <div className="grid gap-px bg-line md:grid-cols-2">
            <motion.article className="bg-surface p-6" variants={fadeUp}>
              <div className="mb-4 flex items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                  <Sparkles className="h-4 w-4 text-lagoon" aria-hidden />
                  <h3 className="text-lg font-semibold text-sea-ink">
                    Claude Code 插件
                  </h3>
                </div>
                <span className="bg-lagoon px-2 py-1 text-[10px] font-bold tracking-wider text-foam">
                  推荐
                </span>
              </div>
              <p className="mb-4 text-sm leading-relaxed text-sea-ink-soft">
                插件同时带来 Lumina 技能和 MCP
                配置。安装后会自动连接当前站点，无需再执行{' '}
                <code>claude mcp add</code>
                。登录控制台生成令牌后，把占位值替换掉即可。
              </p>
              <CopyBlock filename="Claude Code 插件" code={pluginSnippet} />
            </motion.article>

            <motion.article className="bg-chip-bg p-6" variants={fadeUp}>
              <div className="mb-4 flex items-center gap-2">
                <Plug className="h-4 w-4 text-lagoon" aria-hidden />
                <h3 className="text-lg font-semibold text-sea-ink">
                  独立 MCP 配置
                </h3>
              </div>
              <p className="mb-4 text-sm leading-relaxed text-sea-ink-soft">
                适合其他支持 Streamable HTTP 的客户端，也可用于不安装插件的
                Claude Code 环境。下面先以 Cursor 为例。
              </p>
              <CopyBlock filename="~/.cursor/mcp.json" code={cursorSnippet} />
              <p className="mt-3 font-mono text-[11px] break-all text-sea-ink-soft">
                {buildAuthorizationHeader()}
              </p>
            </motion.article>
          </div>

          <motion.div className="mt-5 text-center" variants={fadeUp}>
            <Button
              asChild
              className="bg-sea-ink text-foam hover:bg-lagoon-deep"
            >
              <Link to="/auth/login">
                登录后生成令牌与完整配置
                <ArrowRight className="ml-2 h-4 w-4" aria-hidden />
              </Link>
            </Button>
          </motion.div>
        </motion.div>
      </section>

      <section className="page-wrap px-4 pb-20" aria-label="核心模块">
        <motion.div
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <SectionKicker>核心模块</SectionKicker>
          <motion.h2
            className="display-title mb-8 text-center text-2xl font-bold text-sea-ink sm:text-3xl"
            variants={fadeUp}
          >
            25 个工具，按任务编排
          </motion.h2>
          <div className="mx-auto grid max-w-5xl grid-cols-1 gap-px border border-line bg-line sm:grid-cols-2 lg:grid-cols-3">
            {MCP_TOOL_MODULES.map((mod) => {
              const Icon = iconForModule(mod.id)
              return (
                <motion.article
                  key={mod.id}
                  className="bg-background p-5"
                  aria-label={`${mod.name} 模块说明`}
                  variants={fadeUp}
                >
                  <div className="mb-3 flex h-10 w-10 items-center justify-center border border-line text-lagoon-deep">
                    <Icon className="h-5 w-5" aria-hidden />
                  </div>
                  <div className="mb-2 flex items-baseline justify-between gap-2">
                    <h3 className="text-base font-semibold text-sea-ink">
                      {mod.name}
                    </h3>
                    {mod.note && (
                      <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-lagoon-deep">
                        {mod.note}
                      </span>
                    )}
                  </div>
                  <p className="mb-3 text-sm leading-relaxed text-sea-ink-soft">
                    {mod.summary}
                  </p>
                  <div className="flex flex-wrap gap-1.5">
                    {mod.tools.map((tool) => (
                      <span
                        key={tool.name}
                        className="inline-flex items-center border border-line px-2.5 py-0.5 font-mono text-[10px] font-semibold text-lagoon-deep"
                      >
                        {tool.name}
                      </span>
                    ))}
                  </div>
                </motion.article>
              )
            })}
            <motion.article
              className="bg-background p-5"
              aria-label="Memory 模块说明"
              variants={fadeUp}
            >
              <div className="mb-3 flex h-10 w-10 items-center justify-center border border-line text-lagoon-deep">
                <Brain className="h-5 w-5" aria-hidden />
              </div>
              <div className="mb-2 flex items-baseline justify-between gap-2">
                <h3 className="text-base font-semibold text-sea-ink">Memory</h3>
                <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-sea-ink-soft">
                  设计中
                </span>
              </div>
              <p className="text-sm leading-relaxed text-sea-ink-soft">
                长期决策记忆正在设计中，后续会作为新的工具组加入 Lumina。
              </p>
            </motion.article>
          </div>
        </motion.div>
      </section>

      <section
        className="page-wrap px-4 pb-24 text-center"
        aria-label="行动号召"
      >
        <motion.div
          className={`${shellBase} mx-auto max-w-xl px-8 py-14`}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={sectionStagger}
        >
          <motion.div variants={fadeUp}>
            <Sparkles
              className="mx-auto mb-5 h-10 w-10 text-lagoon"
              aria-hidden
            />
          </motion.div>
          <motion.h2
            className="display-title mb-4 text-2xl font-bold text-sea-ink sm:text-3xl"
            variants={fadeUp}
          >
            服务已经就绪，下一步交给 Lumina
          </motion.h2>
          <motion.p
            className="mb-8 text-sm text-sea-ink-soft"
            variants={fadeUp}
          >
            登录控制台生成接入令牌，插件命令与各客户端配置会自动带入它。
          </motion.p>
          <motion.div
            className="flex flex-col items-center justify-center gap-3 sm:flex-row"
            variants={fadeUp}
          >
            <Button
              asChild
              size="lg"
              className="group relative h-12 bg-sea-ink px-8 text-base font-semibold text-foam hover:bg-lagoon-deep"
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
              className="h-12 px-8 text-base font-medium text-sea-ink-soft transition-colors duration-300 hover:border-sea-ink hover:text-sea-ink"
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

function SectionKicker({ children }: { children: string }) {
  return (
    <motion.div
      className="mb-8 flex items-center justify-center gap-3"
      variants={fadeUp}
    >
      <span className="h-px w-8 bg-line" />
      <span className={kickerBase}>{children}</span>
      <span className="h-px w-8 bg-line" />
    </motion.div>
  )
}
