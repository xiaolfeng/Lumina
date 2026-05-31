import { useState, type FormEvent, useRef } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
  Eye,
  EyeOff,
  Github,
  LogIn,
  BookOpen,
  Brain,
  MessageCircle,
  Sparkles,
  ArrowLeft,
} from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Separator } from '#/components/ui/separator'

export const Route = createFileRoute('/login')({ component: LoginPage })

/* ─── Animation presets ────────────────────────────────── */

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.04 },
  },
}

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.3, ease: 'easeOut' },
  },
}

const cardVariants = {
  hidden: { opacity: 0, scale: 0.94, y: 10 },
  visible: {
    opacity: 1,
    scale: 1,
    y: 0,
    transition: { duration: 0.35, ease: 'easeOut' },
  },
}

/* ─── Data ─────────────────────────────────────────────── */

const highlights = [
  { icon: BookOpen, label: 'RepoWiki' },
  { icon: Brain, label: 'Memory' },
  { icon: MessageCircle, label: 'Q&A' },
] as const

/* ─── Component ────────────────────────────────────────── */

function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [errors, setErrors] = useState<{ email?: string; password?: string }>(
    {},
  )
  const [loading, setLoading] = useState(false)
  const emailInputRef = useRef<HTMLInputElement>(null)
  const passwordInputRef = useRef<HTMLInputElement>(null)

  function validate(): boolean {
    const next: typeof errors = {}
    if (!email.trim()) next.email = '请输入邮箱地址'
    if (!password) next.password = '请输入密码'
    setErrors(next)
    return Object.keys(next).length === 0
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!validate()) {
      if (errors.email) emailInputRef.current?.focus()
      else passwordInputRef.current?.focus()
      return
    }
    setLoading(true)
    console.log({ email, password })
    setTimeout(() => setLoading(false), 1200)
  }

  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* ════════ LEFT: Brand panel ════════ */}
      <motion.aside
        className="relative hidden flex-col justify-between overflow-hidden bg-[var(--sea-ink)] p-10 text-[var(--sand)] lg:flex lg:w-[55%] xl:p-14"
        initial="hidden"
        animate="visible"
        variants={containerVariants}
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
              className="mb-10 inline-flex items-center gap-2 text-sm font-medium text-[var(--lagoon)] transition-colors hover:text-[var(--sand)]"
            >
              <ArrowLeft className="h-4 w-4" aria-hidden />
              返回首页
            </Link>
          </motion.div>

          <motion.div
            className="flex items-center gap-3"
            variants={itemVariants}
          >
            <Sparkles className="h-7 w-7 text-[var(--lagoon)]" aria-hidden />
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
            <span className="text-[var(--lagoon)]">知常曰明</span>
          </motion.h1>

          <motion.p
            className="mb-8 text-base leading-relaxed"
            style={{ color: 'rgba(231, 240, 232, 0.72)' }}
            variants={itemVariants}
          >
            赋予 AI 深度代码认知与长期记忆的知识中枢。
            通过 MCP 协议开放 RepoWiki、Memory、Q&A 三大核心能力，让知识自由流动。
          </motion.p>

          <motion.div
            className="flex flex-wrap gap-3"
            variants={itemVariants}
          >
            {highlights.map((h) => (
              <span
                key={h.label}
                className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium backdrop-blur-sm"
              >
                <h.icon className="h-4 w-4 text-[var(--lagoon)]" aria-hidden />
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

      {/* ════════ RIGHT: Login form ════════ */}
      <main className="relative flex flex-1 items-center justify-center bg-[var(--bg-base)] p-6 lg:w-[45%]">
        {/* Mobile top bar */}
        <div className="absolute left-0 right-0 top-0 flex items-center justify-between p-5 lg:hidden">
          <Link
            to="/"
            className="inline-flex items-center gap-2 text-sm font-medium text-[var(--sea-ink-soft)]"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden />
            返回
          </Link>
          <span className="display-title text-lg font-bold text-[var(--sea-ink)]">
            Lumina
          </span>
        </div>

        <motion.div
          className="island-shell w-full max-w-sm rounded-2xl p-8 lg:max-w-md"
          initial="hidden"
          animate="visible"
          variants={cardVariants}
        >
          {/* Header */}
          <div className="mb-8 text-center">
            <h1 className="display-title text-2xl font-bold text-[var(--sea-ink)] lg:text-3xl">
              欢迎回来
            </h1>
            <p className="mt-2 text-sm text-[var(--sea-ink-soft)]">
              登录到你的知识中枢
            </p>
          </div>

          {/* Form */}
          <form
            onSubmit={handleSubmit}
            noValidate
            className="flex flex-col gap-5"
          >
            {/* Email */}
            <div className="flex flex-col gap-2">
              <Label htmlFor="email" className="text-sm font-medium">
                邮箱
              </Label>
              <Input
                ref={emailInputRef}
                id="email"
                type="email"
                name="email"
                placeholder="you@example.com"
                autoComplete="email"
                spellCheck={false}
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value)
                  if (errors.email)
                    setErrors((prev) => ({ ...prev, email: undefined }))
                }}
                style={{ touchAction: 'manipulation' }}
              />
              {errors.email && (
                <span className="text-xs text-red-500" role="alert">
                  {errors.email}
                </span>
              )}
            </div>

            {/* Password */}
            <div className="flex flex-col gap-2">
              <Label htmlFor="password" className="text-sm font-medium">
                密码
              </Label>
              <div className="relative">
                <Input
                  ref={passwordInputRef}
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  name="password"
                  placeholder="输入密码…"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value)
                    if (errors.password)
                      setErrors((prev) => ({ ...prev, password: undefined }))
                  }}
                  className="pr-10"
                  style={{ touchAction: 'manipulation' }}
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer p-0.5 text-[var(--sea-ink-soft)] transition-colors hover:text-[var(--sea-ink)]"
                  onClick={() => setShowPassword((v) => !v)}
                  aria-label={showPassword ? '隐藏密码' : '显示密码'}
                  tabIndex={-1}
                  style={{ touchAction: 'manipulation' }}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" aria-hidden />
                  ) : (
                    <Eye className="h-4 w-4" aria-hidden />
                  )}
                </button>
              </div>
              {errors.password && (
                <span className="text-xs text-red-500" role="alert">
                  {errors.password}
                </span>
              )}
            </div>

            {/* Forgot password */}
            <div className="text-right">
              <Link
                to="/forgot-password"
                className="text-sm text-[var(--lagoon-deep)] transition-colors hover:text-[var(--lagoon)]"
                style={{ touchAction: 'manipulation' }}
              >
                忘记密码？
              </Link>
            </div>

            {/* Submit */}
            <Button
              type="submit"
              size="lg"
              disabled={loading}
              className="w-full"
              style={{ touchAction: 'manipulation' }}
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <LogIn className="h-4 w-4 animate-pulse" aria-hidden />
                  登录中…
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <LogIn className="h-4 w-4" aria-hidden />
                  登录
                </span>
              )}
            </Button>
          </form>

          {/* Divider */}
          <div className="my-6 flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-xs text-[var(--sea-ink-soft)]">或</span>
            <Separator className="flex-1" />
          </div>

          {/* GitHub */}
          <Button
            variant="outline"
            className="w-full gap-2"
            style={{ touchAction: 'manipulation' }}
          >
            <Github className="h-4 w-4" aria-hidden />
            使用 GitHub 登录
          </Button>

          {/* Register */}
          <p className="mt-6 text-center text-sm text-[var(--sea-ink-soft)]">
            还没有账号？{' '}
            <Link
              to="/register"
              className="font-medium text-[var(--lagoon-deep)] transition-colors hover:text-[var(--lagoon)]"
              style={{ touchAction: 'manipulation' }}
            >
              注册
            </Link>
          </p>
        </motion.div>
      </main>
    </div>
  )
}
