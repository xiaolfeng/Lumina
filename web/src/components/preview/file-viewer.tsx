import { useEffect, useRef, useState } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { ease } from '@lumina/components/motion'

import { PreviewFrame } from '#/components/interact/primitives/preview-frame'
import { formatPreviewSource } from '#/lib/format-preview-source'
import type { PreviewKind } from '#/lib/preview-file'
import { PreviewCodeView } from './code-view'
import { PreviewLpwViewer } from '@lumina/components/lpw'
import { PreviewMdxViewer } from '@lumina/components/mdx'
import { PreviewMarkdownView } from './markdown-view'

export function PreviewFileViewer({
  kind,
  src,
  filename,
}: {
  kind: PreviewKind
  src: string
  filename: string
}) {
  const shouldReduceMotion = useReducedMotion()

  // 对于 HTML/SVG/TSX 视图，复用单一 PreviewFrame 实例，利用内部 motion.div 驱动平滑过渡，
  // 避免外层 key 随文件名切换导致 PreviewFrame 实例被硬销毁重挂并重现全屏 LoadingSlot
  if (kind === 'html' || kind === 'svg' || kind === 'tsx') {
    return <PreviewFrame src={src} className="flex-1" title={filename} />
  }

  return (
    <motion.div
      key={`${filename}:${kind}`}
      initial={{ opacity: shouldReduceMotion ? 1 : 0.4 }}
      animate={{ opacity: 1 }}
      transition={{
        duration: shouldReduceMotion ? 0 : 0.18,
        ease,
      }}
      className="flex min-h-0 flex-1 flex-col h-full w-full"
    >
      {kind === 'lpw' ? (
        <PreviewLpwViewer src={src} filename={filename} />
      ) : kind === 'mdx' ? (
        <PreviewMdxViewer src={src} filename={filename} />
      ) : (
        <PreviewSourceViewer src={src} filename={filename} kind={kind} />
      )}
    </motion.div>
  )
}

function PreviewSourceViewer({
  src,
  filename,
  kind,
}: {
  src: string
  filename: string
  kind: PreviewKind
}) {
  const [source, setSource] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [isUpdating, setIsUpdating] = useState(false)
  const prevFilenameRef = useRef(filename)
  const shouldReduceMotion = useReducedMotion()

  useEffect(() => {
    const ac = new AbortController()
    const fileChanged = prevFilenameRef.current !== filename
    prevFilenameRef.current = filename

    // 仅在真实切换文件时重置 source 触发占位态；同文件增量推送时保持现有视图 (Q-03)
    if (fileChanged) {
      setSource(null)
    }
    setIsUpdating(true)
    setError('')

    void (async () => {
      try {
        const rawUrl = src.includes('?') ? `${src}&raw=1` : `${src}?raw=1`
        const res = await fetch(rawUrl, { signal: ac.signal })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const text = await res.text()
        if (!ac.signal.aborted) {
          setSource(formatPreviewSource(filename, text))
          setIsUpdating(false)
        }
      } catch {
        if (ac.signal.aborted) return
        setError('读取文件内容失败')
        setIsUpdating(false)
      }
    })()
    return () => ac.abort()
  }, [src, filename])

  if (error) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-red-500">{error}</p>
      </div>
    )
  }
  if (source === null) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-sea-ink-soft/50">整理预览中…</p>
      </div>
    )
  }

  return (
    <div className="relative flex min-h-0 flex-1 flex-col h-full w-full">
      <AnimatePresence>
        {isUpdating ? (
          <motion.div
            key="source-sync-bar"
            data-testid="source-sync-bar"
            initial={{ opacity: 0, scaleX: 0 }}
            animate={{ opacity: 1, scaleX: 1 }}
            exit={{ opacity: 0 }}
            transition={{
              duration: shouldReduceMotion ? 0 : 0.15,
              ease,
            }}
            className="absolute top-0 left-0 right-0 z-20 h-0.5 origin-left bg-lagoon shadow-[0_0_8px_var(--color-lagoon)] pointer-events-none"
          />
        ) : null}
      </AnimatePresence>
      {kind === 'markdown' ? (
        <PreviewMarkdownView source={source} />
      ) : (
        <div className="flex min-h-0 flex-1 flex-col">
          <PreviewCodeView filename={filename} source={source} />
        </div>
      )}
    </div>
  )
}
