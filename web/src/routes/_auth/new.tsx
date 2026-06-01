import { useState, useRef } from 'react'
import type { FormEvent } from 'react'
import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { motion } from 'motion/react'
import { Eye, EyeOff, Sparkles } from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'

import { rightItemVariants } from '../_auth'
import { useInitialize } from '#/hooks/useAuth'

export const Route = createFileRoute('/_auth/new')({ component: NewPage })

/* ─── Component ────────────────────────────────────────── */

function NewPage() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [errors, setErrors] = useState<{
    username?: string
    email?: string
    password?: string
    confirmPassword?: string
  }>({})
  const [globalError, setGlobalError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const initMutation = useInitialize()

  const usernameRef = useRef<HTMLInputElement>(null)
  const emailRef = useRef<HTMLInputElement>(null)
  const passwordRef = useRef<HTMLInputElement>(null)
  const confirmPasswordRef = useRef<HTMLInputElement>(null)

  function validate(): boolean {
    const next: typeof errors = {}
    if (!username.trim()) {
      next.username = '请输入用户名'
    } else if (username.trim().length < 3 || username.trim().length > 32) {
      next.username = '用户名长度为 3-32 个字符'
    }

    if (!email.trim()) {
      next.email = '请输入邮箱地址'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      next.email = '请输入有效的邮箱地址'
    }

    if (!password) {
      next.password = '请输入密码'
    } else if (password.length < 6 || password.length > 64) {
      next.password = '密码长度为 6-64 个字符'
    }

    if (!confirmPassword) {
      next.confirmPassword = '请确认密码'
    } else if (confirmPassword !== password) {
      next.confirmPassword = '两次输入的密码不一致'
    }

    setErrors(next)
    return Object.keys(next).length === 0
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setGlobalError(null)
    if (!validate()) {
      if (errors.username) usernameRef.current?.focus()
      else if (errors.email) emailRef.current?.focus()
      else if (errors.password) passwordRef.current?.focus()
      else confirmPasswordRef.current?.focus()
      return
    }
    initMutation.mutate(
      {
        username: username.trim(),
        email: email.trim(),
        password,
      },
      {
        onSuccess: async () => {
          await queryClient.invalidateQueries({ queryKey: ['auth', 'status'] })
          await router.invalidate()
          setSuccess(true)
          setTimeout(() => {
            router.navigate({ to: '/login' })
          }, 1500)
        },
        onError: (err) => {
          setGlobalError(err.message)
        },
      },
    )
  }

  return (
    <>
      {/* Header */}
      <motion.div className="mb-8 text-center" variants={rightItemVariants}>
        <h1 className="display-title text-2xl font-bold text-[var(--sea-ink)] lg:text-3xl">
          初始化系统
        </h1>
        <p className="mt-2 text-sm text-[var(--sea-ink-soft)]">
          创建首个管理员账户以开启 Lumina
        </p>
      </motion.div>

      {/* Form */}
      <motion.form
        variants={rightItemVariants}
        onSubmit={handleSubmit}
        noValidate
        className="flex flex-col gap-5"
      >
        {/* Username */}
        <div className="flex flex-col gap-2">
          <Label htmlFor="username" className="text-sm font-medium">
            用户名
          </Label>
          <Input
            ref={usernameRef}
            id="username"
            type="text"
            name="username"
            placeholder="输入用户名…"
            autoComplete="username"
            spellCheck={false}
            value={username}
            onChange={(e) => {
              setUsername(e.target.value)
              if (errors.username)
                setErrors((prev) => ({ ...prev, username: undefined }))
            }}
            style={{ touchAction: 'manipulation' }}
          />
          {errors.username && (
            <span className="text-xs text-red-500" role="alert">
              {errors.username}
            </span>
          )}
        </div>

        {/* Email */}
        <div className="flex flex-col gap-2">
          <Label htmlFor="email" className="text-sm font-medium">
            邮箱
          </Label>
          <Input
            ref={emailRef}
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
              ref={passwordRef}
              id="password"
              type={showPassword ? 'text' : 'password'}
              name="password"
              placeholder="输入密码…"
              autoComplete="new-password"
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

        {/* Confirm Password */}
        <div className="flex flex-col gap-2">
          <Label htmlFor="confirmPassword" className="text-sm font-medium">
            确认密码
          </Label>
          <div className="relative">
            <Input
              ref={confirmPasswordRef}
              id="confirmPassword"
              type={showConfirmPassword ? 'text' : 'password'}
              name="confirmPassword"
              placeholder="再次输入密码…"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value)
                if (errors.confirmPassword)
                  setErrors((prev) => ({
                    ...prev,
                    confirmPassword: undefined,
                  }))
              }}
              className="pr-10"
              style={{ touchAction: 'manipulation' }}
            />
            <button
              type="button"
              className="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer p-0.5 text-[var(--sea-ink-soft)] transition-colors hover:text-[var(--sea-ink)]"
              onClick={() => setShowConfirmPassword((v) => !v)}
              aria-label={showConfirmPassword ? '隐藏密码' : '显示密码'}
              tabIndex={-1}
              style={{ touchAction: 'manipulation' }}
            >
              {showConfirmPassword ? (
                <EyeOff className="h-4 w-4" aria-hidden />
              ) : (
                <Eye className="h-4 w-4" aria-hidden />
              )}
            </button>
          </div>
          {errors.confirmPassword && (
            <span className="text-xs text-red-500" role="alert">
              {errors.confirmPassword}
            </span>
          )}
        </div>

        {/* Login link */}
        <div className="text-right">
          <Link
            to="/login"
            className="text-sm text-[var(--lagoon-deep)] transition-colors hover:text-[var(--lagoon)]"
            style={{ touchAction: 'manipulation' }}
          >
            返回登录
          </Link>
        </div>

        {/* Submit */}
        <Button
          type="submit"
          size="lg"
          disabled={initMutation.isPending || success}
          className="w-full"
          style={{ touchAction: 'manipulation' }}
        >
          {initMutation.isPending ? (
            <span className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 animate-pulse" aria-hidden />
              初始化中…
            </span>
          ) : success ? (
            <span className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" aria-hidden />
              初始化成功
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" aria-hidden />
              创建管理员账户
            </span>
          )}
        </Button>

        {/* Global error */}
        {globalError && (
          <span className="text-center text-xs text-red-500" role="alert">
            {globalError}
          </span>
        )}
      </motion.form>
    </>
  )
}
