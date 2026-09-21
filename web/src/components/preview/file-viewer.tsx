import { useEffect, useState } from 'react'

import { PreviewFrame } from '#/components/interact/primitives/preview-frame'
import { formatPreviewSource } from '#/lib/format-preview-source'
import type { PreviewKind } from '#/lib/preview-file'
import { PreviewCodeView } from './code-view'
import { PreviewLpwViewer } from '@lumina/components/lpw'
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
  if (kind === 'lpw') {
    return <PreviewLpwViewer src={src} filename={filename} />
  }

  if (kind === 'html' || kind === 'svg' || kind === 'tsx') {
    return <PreviewFrame src={src} className="flex-1" title={filename} />
  }

  return <PreviewSourceViewer src={src} filename={filename} kind={kind} />
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

  useEffect(() => {
    const ac = new AbortController()
    setSource(null)
    setError('')
    void (async () => {
      try {
        const rawUrl = src.includes('?') ? `${src}&raw=1` : `${src}?raw=1`
        const res = await fetch(rawUrl, { signal: ac.signal })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const text = await res.text()
        setSource(formatPreviewSource(filename, text))
      } catch {
        if (ac.signal.aborted) return
        setError('读取文件内容失败')
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
  if (kind === 'markdown') {
    return <PreviewMarkdownView source={source} />
  }
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <PreviewCodeView filename={filename} source={source} />
    </div>
  )
}
