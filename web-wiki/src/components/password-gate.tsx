/**
 * PasswordGate — 授权守卫组件
 *
 * 根据授权状态决定渲染 children 或 PasswordInput。
 *
 * 三种状态：
 * 1. loading → 骨架屏占位
 * 2. error → 错误提示 + 重试按钮
 * 3. 已就绪 → 无需密码 / 已授权 → 渲染 children；需要密码且未授权 → PasswordInput
 */
import type { ReactNode } from 'react'
import { AlertCircle, RefreshCw, Loader2 } from 'lucide-react'
import { useWikiAuthCheck } from '#/hooks/useWikiAuth'
import { PasswordInput } from './password-input'

interface PasswordGateProps {
  wikiId: string
  children: ReactNode
}

export function PasswordGate({ wikiId, children }: PasswordGateProps) {
  const { data, isLoading, isError, error, refetch, isFetching } =
    useWikiAuthCheck(wikiId)

  // 加载中 → 骨架屏
  if (isLoading) {
    return <PasswordGateSkeleton />
  }

  // 错误状态（非首次加载时的错误）
  if (isError) {
    return (
      <PasswordGateError
        error={error}
        onRetry={() => refetch()}
        isRetrying={isFetching}
      />
    )
  }

  // 不需要密码 → 直接展示
  if (!data?.password_required) {
    return <>{children}</>
  }

  // 已授权 → 展示
  if (data.authenticated) {
    return <>{children}</>
  }

  // 需要密码且未授权 → 显示密码输入
  return <PasswordInput wikiId={wikiId} />
}

// ── 子组件：骨架屏 ──

function PasswordGateSkeleton() {
  return (
    <div className="flex min-h-[400px] items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm space-y-6 rounded-xl border border-line bg-surface-strong p-6 shadow-sm">
        {/* 图标骨架 */}
        <div className="flex justify-center">
          <div className="h-14 w-14 animate-pulse rounded-full bg-muted" />
        </div>
        {/* 标题骨架 */}
        <div className="space-y-2 text-center">
          <div className="mx-auto h-6 w-32 animate-pulse rounded bg-muted" />
          <div className="mx-auto h-4 w-48 animate-pulse rounded bg-muted" />
        </div>
        {/* 输入框骨架 */}
        <div className="space-y-2">
          <div className="h-4 w-16 animate-pulse rounded bg-muted" />
          <div className="h-10 w-full animate-pulse rounded-md bg-muted" />
        </div>
        {/* 按钮骨架 */}
        <div className="h-10 w-full animate-pulse rounded-md bg-muted" />
      </div>
    </div>
  )
}

// ── 子组件：错误提示 ──

interface PasswordGateErrorProps {
  error: Error | null
  onRetry: () => void
  isRetrying: boolean
}

function PasswordGateError({
  error,
  onRetry,
  isRetrying,
}: PasswordGateErrorProps) {
  return (
    <div className="flex min-h-[400px] items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm space-y-4 rounded-xl border border-destructive/20 bg-surface-strong p-6 text-center shadow-sm">
        <AlertCircle className="mx-auto h-10 w-10 text-destructive" />
        <h3 className="text-lg font-semibold text-sea-ink">加载失败</h3>
        <p className="text-sm text-sea-ink-soft">
          {error?.message || '无法连接到服务器，请检查网络后重试'}
        </p>
        <button
          onClick={onRetry}
          disabled={isRetrying}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {isRetrying ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              重试中...
            </>
          ) : (
            <>
              <RefreshCw className="h-4 w-4" />
              重试
            </>
          )}
        </button>
      </div>
    </div>
  )
}
