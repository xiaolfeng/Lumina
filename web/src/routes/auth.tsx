import { createFileRoute, Link, Outlet } from '@tanstack/react-router'
import { motion, type Variants } from 'motion/react'
import {
  BookOpen,
  Brain,
  MessageCircle,
  Sparkles,
  ArrowLeft,
} from 'lucide-react'

export const Route = createFileRoute('/auth')({
  component: AuthLayout,
})

/* ─── Animation presets ────────────────────────────────── */

const ease: [number, number, number, number] = [0.16, 1, 0.3, 1]

const leftPanelVariants: Variants = {
  hidden: { opacity: 0, x: -10 },
  visible: {
    opacity: 1,
    x: 0,
    transition: {
      duration: 0.6,
      ease,
      staggerChildren: 0.06,
      delayChildren: 0.1,
    },
  },
}

const rightContainerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      duration: 0.5,
      ease,
      staggerChildren: 0.06,
      delayChildren: 0.35,
    },
  },
}

const itemVariants: Variants = {
  hidden: { opacity: 0, x: -60 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.5, ease },
  },
}

const rightItemVariants: Variants = {
  hidden: { opacity: 0, x: 60 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.5, ease },
  },
}

/* ─── Data ─────────────────────────────────────────────── */

const highlights = [
  { icon: BookOpen, label: 'RepoWiki' },
  { icon: Brain, label: 'Memory' },
  { icon: MessageCircle, label: 'Q&A' },
] as const

/* ─── Layout Component ─────────────────────────────────── */

function AuthLayout() {
  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* ════════ LEFT: Brand panel ════════ */}
      <motion.aside
        className="relative hidden flex-col justify-between overflow-hidden bg-sea-ink p-10 text-sand lg:flex lg:w-[55%] xl:p-14 dark:bg-[#1a1512] dark:text-[#ede5d8]"
        initial="hidden"
        animate="visible"
        variants={leftPanelVariants}
      >
        {/* Decorative orbs */}
        <div
          className="pointer-events-none absolute -left-24 -top-24 h-72 w-72 rounded-full opacity-20 blur-3xl"
          style={{ background: 'var(--lagoon)' }}
        />
        <div
          className="pointer-events-none absolute -bottom-32 -right-32 h-96 w-96 rounded-full opacity-15 blur-3xl"
          style={{ background: 'var(--palm)' }}
        />

        {/* Top */}
        <div className="relative z-10">
          <motion.div variants={itemVariants}>
            <Link
              to="/"
              className="mb-10 inline-flex items-center gap-2 text-sm font-medium text-lagoon transition-colors hover:text-sand dark:hover:text-[#ede5d8]"
            >
              <ArrowLeft className="h-4 w-4" aria-hidden />
              返回首页
            </Link>
          </motion.div>

          <motion.div
            className="flex items-center gap-3"
            variants={itemVariants}
          >
            <Sparkles className="h-7 w-7 text-lagoon" aria-hidden />
            <span className="display-title text-2xl font-bold tracking-tight">
              Lumina · 微明
            </span>
          </motion.div>
        </div>

        {/* Middle */}
        <div className="relative z-10 max-w-md">
          <motion.h1
            className="display-title mb-4 text-4xl font-bold leading-tight tracking-tight xl:text-5xl"
            variants={itemVariants}
          >
            烛照幽微
            <br />
            <span className="text-lagoon">知常曰明</span>
          </motion.h1>

          <motion.p
            className="mb-8 text-base leading-relaxed"
            style={{ color: 'rgba(231, 240, 232, 0.72)' }}
            variants={itemVariants}
          >
            赋予 AI 深度代码认知与长期记忆的知识中枢。 通过 MCP 协议开放
            RepoWiki、Memory、Q&A 三大核心能力，让知识自由流动。
          </motion.p>

          <motion.div className="flex flex-wrap gap-3" variants={itemVariants}>
            {highlights.map((h) => (
              <span
                key={h.label}
                className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium backdrop-blur-sm"
              >
                <h.icon className="h-4 w-4 text-lagoon" aria-hidden />
                {h.label}
              </span>
            ))}
          </motion.div>
        </div>

        {/* Bottom */}
        <motion.p
          className="relative z-10 text-xs text-white/30"
          variants={itemVariants}
        >
          © 2026 Xiao Lfeng (筱锋) · MIT License
        </motion.p>
      </motion.aside>

      {/* ════════ RIGHT: Content area ════════ */}
      <main className="relative flex flex-1 items-center justify-center bg-bg-base p-6 lg:w-[45%]">
        {/* Mobile top bar */}
        <div className="absolute left-0 right-0 top-0 flex items-center justify-between p-5 lg:hidden">
          <Link
            to="/"
            className="inline-flex items-center gap-2 text-sm font-medium text-sea-ink-soft"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden />
            返回
          </Link>
          <span className="display-title text-lg font-bold text-sea-ink">
            Lumina
          </span>
        </div>

        <motion.div
          className="w-full max-w-sm lg:max-w-md"
          initial="hidden"
          animate="visible"
          variants={rightContainerVariants}
        >
          <Outlet />
        </motion.div>
      </main>
    </div>
  )
}

/* ─── Re-export variants for child routes ──────────────── */

export { rightItemVariants }
