import { useEffect, useRef, useState } from 'react'

import { PreviewLpwInlineViewer } from '@lumina/components/lpw'
import { PreviewMdxInlineViewer } from '@lumina/components/mdx'
import { getPreviewFileByID } from '#/lib/apis/preview'
import { previewKindFromFilename } from '#/lib/preview-file'

/**
 * PreviewFrame — 跨文件引用预览 iframe（src 模式）
 *
 * 与 SandboxFrame（srcDoc 内联单片段）不同，本组件让 iframe 的 src 直接指向
 * 后端 serve 接口（/api/v1/preview/sessions/{hash}/files/{filename}），使浏览器
 * 原生解析 <link href="style.css"> / <script src="app.js"> 等相对引用 ——
 * 相对引用会基于 iframe 的 src 解析，命中同 Session 的其它文件。
 *
 * - sandbox="allow-scripts allow-popups" 刻意不配 allow-same-origin，脚本运行在隔离 origin
 * - iframe 高度由父容器决定（h-full），内部滚动，无需 postMessage 高度回传
 */
export interface PreviewFrameProps {
  /** 指向 serve 接口的文件 URL（支持跨文件相对引用） */
  src: string
  /** 作用于外层容器的 className（布局） */
  className?: string
  /** iframe 无障碍标题 */
  title?: string
}

function LoadingSlot() {
  return (
    <div className="absolute inset-0 z-10 flex items-center justify-center bg-bg-base">
      <p className="text-sm text-sea-ink-soft/50">正在加载中</p>
    </div>
  )
}

export function PreviewFrame({ src, className, title }: PreviewFrameProps) {
  const [mounted, setMounted] = useState(false)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    setLoaded(false)
  }, [src])

  return (
    <div className={`relative min-h-0 h-full w-full ${className ?? ''}`}>
      {(!mounted || !loaded) && <LoadingSlot />}
      {mounted ? (
        <iframe
          src={src}
          title={title ?? '前端预览'}
          sandbox="allow-scripts allow-popups"
          onLoad={() => setLoaded(true)}
          className="block h-full w-full border-0"
        />
      ) : null}
    </div>
  )
}

/**
 * PreviewSupplement — 解析 supplement content_type=preview 的引用并渲染
 *
 * content 为 JSON：{"session_id":"...","file_id":"..."}。本组件通过 file_id
 * 反查文件详情（含 session_hash 与 filename），再构造 serve 地址交给 PreviewFrame。
 */
export function PreviewSupplement({ content }: { content: string }) {
  const [src, setSrc] = useState('')
  const [filename, setFilename] = useState('')
  const [error, setError] = useState('')
  const cancelledRef = useRef(false)

  useEffect(() => {
    cancelledRef.current = false
    setSrc('')
    setFilename('')
    setError('')

    void (async () => {
      let fileId = ''
      try {
        const parsed: unknown = JSON.parse(content)
        if (parsed !== null && typeof parsed === 'object') {
          const raw = (parsed as Record<string, unknown>).file_id
          if (typeof raw === 'string') fileId = raw
        }
      } catch {
        // 非 JSON 内容，fileId 保持空
      }

      if (fileId === '') {
        if (!cancelledRef.current) setError('无效的预览引用')
        return
      }

      try {
        const res = await getPreviewFileByID(fileId)
        if (cancelledRef.current) return
        const detail = res.data
        if (detail) {
          // filename 需编码（含 #/? 时原样拼接会 404）；lumina_frame=1 标记为
          // 非顶层文档，后端直出文件内容而不回落 SPA
          const finalFilename =
            detail.filename || (detail as any).file_name || ''
          const finalHash =
            detail.session_hash || (detail as any).sessionHash || ''
          setFilename(finalFilename)
          setSrc(
            `/preview/${finalHash}/${encodeURIComponent(
              finalFilename,
            )}?lumina_frame=1`,
          )
        } else {
          setError('预览文件不存在')
        }
      } catch {
        if (!cancelledRef.current) setError('解析预览引用失败')
      }
    })()

    return () => {
      cancelledRef.current = true
    }
  }, [content])

  if (error) {
    return (
      <div className="flex h-full min-h-80 w-full items-center justify-center">
        <p className="text-xs text-red-500">{error}</p>
      </div>
    )
  }
  if (src === '') {
    return (
      <div className="relative h-full min-h-80 w-full">
        <LoadingSlot />
      </div>
    )
  }
  if (previewKindFromFilename(filename) === 'lpw') {
    return <PreviewLpwInlineViewer src={src} filename={filename} />
  }
  if (previewKindFromFilename(filename) === 'mdx') {
    return <PreviewMdxInlineViewer src={src} filename={filename} />
  }
  return <PreviewFrame src={src} className="min-h-80" />
}
