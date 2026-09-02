import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { Loader2, X } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'

const WARN_THRESHOLD = 30_000
const DANGER_THRESHOLD = 120_000

interface SupplementLoadingBannerProps {
  onDismiss: () => void
}

export function SupplementLoadingBanner({
  onDismiss,
}: SupplementLoadingBannerProps) {
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    const start = Date.now()
    const timer = setInterval(() => {
      setElapsed(Date.now() - start)
    }, 1000)
    return () => clearInterval(timer)
  }, [])

  const isDanger = elapsed >= DANGER_THRESHOLD
  const showHint = elapsed >= WARN_THRESHOLD

  return (
    <div
      className={`flex items-center gap-2 border px-3 py-2.5 text-sm transition-colors duration-300 ${
        isDanger
          ? 'border-destructive/30 bg-destructive/8 text-destructive'
          : showHint
            ? 'border-lagoon/35 bg-lagoon/8 text-lagoon-deep'
            : 'border-line bg-chip-bg text-sea-ink-soft'
      }`}
      role="status"
      aria-live="polite"
    >
      <Loader2 className="size-4 shrink-0 animate-spin" aria-hidden />
      <div className="min-w-0 flex-1">
        <p>
          {isDanger
            ? '已等待较长时间，若 AI Agent 行为中断请让其继续或忽略补充'
            : showHint
              ? '请耐心等待或检查 AI Agent 是否出现中断'
              : '正在加载补充内容…'}
        </p>
      </div>
      {showHint && (
        <Button
          variant="ghost"
          size="sm"
          onClick={onDismiss}
          className={`shrink-0 text-xs ${
            isDanger
              ? 'text-destructive hover:bg-destructive/10'
              : 'text-lagoon-deep hover:bg-lagoon/10'
          }`}
        >
          <X className="mr-1 size-3.5" aria-hidden />
          忽略
        </Button>
      )}
    </div>
  )
}

/** 补充等待：提示条 + 控件区灰度遮罩。所有题型共用，避免各组件各自实现。 */
export function SupplementWaitOverlay({
  loading,
  onDismiss,
  children,
}: {
  loading: boolean
  onDismiss?: () => void
  children: ReactNode
}) {
  return (
    <div className="space-y-3">
      {loading && <SupplementLoadingBanner onDismiss={() => onDismiss?.()} />}
      <div
        className="relative"
        aria-busy={loading || undefined}
        inert={loading}
      >
        {loading && (
          <div className="absolute inset-0 z-10 bg-bg-base/55" aria-hidden />
        )}
        <div
          className={
            loading ? 'pointer-events-none select-none grayscale' : undefined
          }
        >
          {children}
        </div>
      </div>
    </div>
  )
}
