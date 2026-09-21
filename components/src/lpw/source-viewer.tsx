import { useCallback, useEffect, useState } from 'react'
import { LpwDocumentViewer } from './document-viewer'
import { LpwRuntimeProvider } from './runtime-provider'
import type { LpwRuntimeConfig } from './runtime-provider'

export interface UseLpwSourceResult {
  source: string | null
  loading: boolean
  error: string | null
  reload: () => void
}

export function useLpwSource(src: string): UseLpwSourceResult {
  const [source, setSource] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [fetchIndex, setFetchIndex] = useState(0)

  const reload = useCallback(() => {
    setFetchIndex((i) => i + 1)
  }, [])

  useEffect(() => {
    const ac = new AbortController()
    setLoading(true)
    setError(null)

    void (async () => {
      try {
        const res = await fetch(src, { signal: ac.signal })
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`)
        }
        const text = await res.text()
        if (!ac.signal.aborted) {
          setSource(text)
          setLoading(false)
        }
      } catch (err) {
        if (ac.signal.aborted) return
        setError(err instanceof Error ? err.message : '读取 LPW 文件失败')
        setLoading(false)
      }
    })()

    return () => ac.abort()
  }, [src, fetchIndex])

  return { source, loading, error, reload }
}

export interface LpwSourceViewerProps {
  src: string
  filename: string
  variant?: 'page' | 'inline'
  runtimeConfig?: Omit<LpwRuntimeConfig, 'assetBaseUrl'>
}

export function LpwSourceViewer({
  src,
  filename: _filename,
  variant = 'page',
  runtimeConfig,
}: LpwSourceViewerProps) {
  // 1. assetBaseUrl = src 去掉最后一段（'/preview/h/index.lpw' → '/preview/h'）
  const lastSlashIndex = src.lastIndexOf('/')
  const assetBaseUrl =
    lastSlashIndex !== -1 ? src.slice(0, lastSlashIndex) : ''

  // 2. 异步拉取 source
  const { source, loading, error, reload } = useLpwSource(src)

  if (loading) {
    if (variant === 'inline') {
      return (
        <div className="flex min-h-80 w-full items-center justify-center border border-line bg-surface-muted/20 p-8">
          <p className="text-xs text-sea-ink-soft/60">加载中…</p>
        </div>
      )
    }
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <p className="text-xs text-sea-ink-soft/60">加载中…</p>
      </div>
    )
  }

  if (error || source === null) {
    if (variant === 'inline') {
      return (
        <div className="flex min-h-80 w-full flex-col items-center justify-center gap-2 border border-line bg-surface p-8 text-xs text-red-600">
          <p>读取 LPW 文件失败: {error || '未知错误'}</p>
          <button
            type="button"
            onClick={reload}
            className="cursor-pointer border border-line bg-surface px-3 py-1 text-xs text-sea-ink hover:bg-surface-muted"
          >
            重试
          </button>
        </div>
      )
    }
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-xs text-red-600">
        <p>读取 LPW 文件失败: {error || '未知错误'}</p>
        <button
          type="button"
          onClick={reload}
          className="cursor-pointer border border-line bg-surface px-3 py-1 text-xs text-sea-ink hover:bg-surface-muted"
        >
          重试
        </button>
      </div>
    )
  }

  const effectiveConfig: LpwRuntimeConfig = {
    ...runtimeConfig,
    assetBaseUrl,
  }

  if (variant === 'inline') {
    return (
      <div className="min-h-80 w-full border border-line bg-surface">
        <LpwRuntimeProvider config={effectiveConfig}>
          <LpwDocumentViewer source={source} />
        </LpwRuntimeProvider>
      </div>
    )
  }

  return (
    <div className="h-full w-full overflow-y-auto">
      <LpwRuntimeProvider config={effectiveConfig}>
        <LpwDocumentViewer source={source} />
      </LpwRuntimeProvider>
    </div>
  )
}

export function PreviewLpwViewer(props: { src: string; filename: string }) {
  return <LpwSourceViewer {...props} variant="page" />
}

export function PreviewLpwInlineViewer(props: {
  src: string
  filename: string
}) {
  return <LpwSourceViewer {...props} variant="inline" />
}
