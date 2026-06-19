import { useState, useRef, useEffect } from 'react'
import type { FormEvent } from 'react'
import { createFileRoute, Link, useRouter, useSearch } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Eye, EyeOff, Fingerprint, LogIn } from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Separator } from '#/components/ui/separator'

import { useLogin, useAuth, useBiometricAvailability, useBiometricLogin } from '#/hooks/useAuth'
import { getSafeRedirect } from '#/lib/apis/client'
import { rightItemVariants } from '../auth'

interface LoginSearch {
  redirect?: string
}

export const Route = createFileRoute('/auth/login')({
  validateSearch: (search: Record<string, unknown>): LoginSearch => {
    return {
      redirect:
        typeof search.redirect === 'string' ? search.redirect : undefined,
    }
  },
  component: LoginPage,
})

/* ─── Component ────────────────────────────────────────── */

function LoginPage() {
  const router = useRouter()
  const search = useSearch({ from: '/auth/login' })
  const login = useLogin()
  const biometricAvail = useBiometricAvailability()
  const biometricLogin = useBiometricLogin()
  const { isAuthenticated } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [errors, setErrors] = useState<{ email?: string; password?: string }>(
    {},
  )
  const [globalError, setGlobalError] = useState<string | null>(null)
  const emailInputRef = useRef<HTMLInputElement>(null)
  const passwordInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isAuthenticated) {
      const target = getSafeRedirect(search.redirect, '/console/dashboard')
      void router.navigate({ to: target })
    }
  }, [isAuthenticated, search.redirect, router])

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
    setGlobalError(null)
    login.mutate(
      { account: email, password },
      {
        onSuccess: () => {
          const target = getSafeRedirect(search.redirect, '/console/dashboard')
          void router.navigate({ to: target })
        },
        onError: (error) => {
          setGlobalError(error.message)
        },
      },
    )
  }

  return (
    <>
      {/* Header */}
      <motion.div className="mb-8 text-center" variants={rightItemVariants}>
        <h1 className="display-title text-2xl font-bold text-sea-ink lg:text-3xl">
          欢迎回来
        </h1>
        <p className="mt-2 text-sm text-sea-ink-soft">
          登录到你的知识中枢
        </p>
      </motion.div>

      {/* Form */}
      <motion.form
        variants={rightItemVariants}
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
              className="absolute right-3 top-1/2 -translate-y-1/2 cursor-pointer p-0.5 text-sea-ink-soft transition-colors hover:text-sea-ink"
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

        {/* Reset password hint */}
        <div className="text-right">
          <Link
            to="/auth/reset-password"
            className="text-sm text-lagoon-deep transition-colors hover:text-lagoon"
            style={{ touchAction: 'manipulation' }}
          >
            忘记密码？
          </Link>
        </div>

        {/* Submit */}
        <Button
          type="submit"
          size="lg"
          disabled={login.isPending}
          className="w-full"
          style={{ touchAction: 'manipulation' }}
        >
          {login.isPending ? (
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

        {globalError && (
          <span className="text-xs text-red-500" role="alert">
            {globalError}
          </span>
        )}
      </motion.form>

      {/* 生物特征登录（仅当已注册凭证时显示） */}
      {biometricAvail?.data?.data?.is_available === true && (
        <>
          {/* Divider */}
          <motion.div
            className="my-6 flex items-center gap-3"
            variants={rightItemVariants}
          >
            <Separator className="flex-1" />
            <span className="text-xs text-sea-ink-soft">或</span>
            <Separator className="flex-1" />
          </motion.div>

          {/* Biometric Login Button */}
          <motion.div variants={rightItemVariants}>
            <Button
              variant="outline"
              className="w-full gap-2"
              onClick={() => biometricLogin.mutate()}
              disabled={biometricLogin.isPending}
              style={{ touchAction: 'manipulation' }}
            >
              <Fingerprint className="h-4 w-4" aria-hidden />
              {biometricLogin.isPending ? '认证中…' : '使用生物特征登录'}
            </Button>
          </motion.div>
        </>
      )}
    </>
  )
}
