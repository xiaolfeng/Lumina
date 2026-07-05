/**
 * PasswordInput — 密码输入表单
 *
 * 内容区内联的密码输入卡片，符合文档站视觉风格。
 * 特性：
 * - 居中卡片布局
 * - 密码显示/隐藏切换
 * - Enter 键提交
 * - 加载状态禁用
 * - 错误提示
 */
import { useState } from 'react'
import type { FormEvent } from 'react'
import { Lock, Eye, EyeOff, Loader2 } from 'lucide-react'
import { useWikiAuth } from '#/hooks/useWikiAuth'

interface PasswordInputProps {
  wikiId: string
}

export function PasswordInput({ wikiId }: PasswordInputProps) {
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const authMutation = useWikiAuth(wikiId)

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!password.trim() || authMutation.isPending) return
    authMutation.mutate(password)
  }

  return (
    <div className="flex min-h-[400px] items-center justify-center px-4 py-12">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm space-y-6 rounded-xl border border-line bg-surface-strong p-6 shadow-sm"
      >
        {/* 图标 + 标题 */}
        <div className="text-center">
          <div className="mx-auto mb-4 inline-flex h-14 w-14 items-center justify-center rounded-full bg-lagoon/10">
            <Lock className="h-7 w-7 text-lagoon" />
          </div>
          <h2 className="display-title text-xl font-semibold text-sea-ink">
            需要授权访问
          </h2>
          <p className="mt-2 text-sm leading-relaxed text-sea-ink-soft">
            此 Wiki 需要密码才能查看内容
          </p>
        </div>

        {/* 密码输入框 */}
        <div className="space-y-2">
          <label
            htmlFor="wiki-password"
            className="block text-sm font-medium text-sea-ink"
          >
            访问密码
          </label>
          <div className="relative">
            <input
              id="wiki-password"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入访问密码"
              autoComplete="current-password"
              disabled={authMutation.isPending}
              autoFocus
              className="w-full rounded-md border border-input bg-surface px-3 py-2.5 pr-10 text-sm text-foreground placeholder:text-muted-foreground focus:border-ring focus:outline-none focus:ring-2 focus:ring-ring/20 disabled:cursor-not-allowed disabled:opacity-50"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              tabIndex={-1}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
              aria-label={showPassword ? '隐藏密码' : '显示密码'}
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        {/* 错误提示 */}
        {authMutation.isError && (
          <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {authMutation.error.message || '密码错误，请重试'}
          </p>
        )}

        {/* 提交按钮 */}
        <button
          type="submit"
          disabled={authMutation.isPending || !password.trim()}
          className="flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:pointer-events-none disabled:opacity-50"
        >
          {authMutation.isPending ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              验证中...
            </>
          ) : (
            '确认访问'
          )}
        </button>
      </form>
    </div>
  )
}
