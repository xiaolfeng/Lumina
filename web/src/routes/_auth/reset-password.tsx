import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Terminal, ArrowLeft, Container } from 'lucide-react'

import { rightItemVariants } from '../_auth'

export const Route = createFileRoute('/_auth/reset-password')({
  component: ResetPasswordPage,
})

function ResetPasswordPage() {
  return (
    <>
      {/* Header */}
      <motion.div className="mb-8 text-center" variants={rightItemVariants}>
        <h1 className="display-title text-2xl font-bold text-[var(--sea-ink)] lg:text-3xl">
          重置密码
        </h1>
        <p className="mt-2 text-sm text-[var(--sea-ink-soft)]">
          本系统不提供在线密码重置功能
        </p>
      </motion.div>

      {/* Instructions */}
      <motion.div
        variants={rightItemVariants}
        className="flex flex-col gap-5"
      >
        <div className="rounded-xl border border-[var(--line)] bg-card p-5">
          <div className="mb-3 flex items-center gap-2 text-sm font-medium text-[var(--sea-ink)]">
            <Terminal className="h-4 w-4 text-[var(--lagoon)]" aria-hidden />
            终端方式
          </div>
          <p className="mb-2 text-sm text-[var(--sea-ink-soft)]">
            请在终端中运行以下命令：
          </p>
          <code className="block rounded-md bg-[var(--sea-ink)] px-4 py-3 text-sm font-mono text-[var(--sand)] dark:bg-[var(--sand)] dark:text-[var(--sea-ink)]">
            lumina reset-user
          </code>
        </div>

        <div className="rounded-xl border border-[var(--line)] bg-card p-5">
          <div className="mb-3 flex items-center gap-2 text-sm font-medium text-[var(--sea-ink)]">
            <Container className="h-4 w-4 text-[var(--lagoon)]" aria-hidden />
            Docker 方式
          </div>
          <p className="mb-2 text-sm text-[var(--sea-ink-soft)]">
            若使用 Docker 部署，请进入容器内部执行：
          </p>
          <code className="block rounded-md bg-[var(--sea-ink)] px-4 py-3 text-sm font-mono text-[var(--sand)] dark:bg-[var(--sand)] dark:text-[var(--sea-ink)]">
            docker exec -it {'<container>'} lumina reset-user
          </code>
        </div>
      </motion.div>

      {/* Back to login */}
      <motion.p
        className="mt-6 text-center text-sm text-[var(--sea-ink-soft)]"
        variants={rightItemVariants}
      >
        <Link
          to="/login"
          className="inline-flex items-center gap-1 font-medium text-[var(--lagoon-deep)] transition-colors hover:text-[var(--lagoon)]"
          style={{ touchAction: 'manipulation' }}
        >
          <ArrowLeft className="h-3 w-3" aria-hidden />
          返回登录
        </Link>
      </motion.p>
    </>
  )
}
