import { useEffect, useRef, useState } from 'react'

import { getPreviewFileByID } from '#/lib/apis/preview'

/**
 * PreviewFrame — 跨文件引用预览 iframe（src 模式）
 *
 * 与 SandboxFrame（srcDoc 内联单片段）不同，本组件让 iframe 的 src 直接指向
 * 后端 serve 接口（/api/v1/preview/sessions/{hash}/files/{filename}），使浏览器
 * 原生解析 <link href="style.css"> / <script src="app.js"> 等相对引用 ——
 * 相对引用会基于 iframe 的 src 解析，命中同 Session 的其它文件。
 *
 * - sandbox="allow-scripts" 刻意不配 allow-same-origin，脚本运行在隔离 origin
 * - iframe 高度由父容器决定（h-full），内部滚动，无需 postMessage 高度回传
 */
export interface PreviewFrameProps {
  /** 指向 serve 接口的文件 URL（支持跨文件相对引用） */
  src: string
  /** 作用于 iframe 元素的 className（布局） */
  className?: string
  /** iframe 无障碍标题 */
  title?: string
}

export function PreviewFrame({ src, className, title }: PreviewFrameProps) {
  const [mounted, setMounted] = useState(false)

  // 首次挂载后再渲染 iframe，规避 SSR 下 src 序列化问题
  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return <div className={className} />
  }

  return (
    <iframe
      src={src}
      title={title ?? '前端预览'}
      sandbox="allow-scripts"
      className={`block h-full w-full border-0 ${className ?? ''}`}
    />
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
  const [error, setError] = useState('')
  const cancelledRef = useRef(false)

  useEffect(() => {
    cancelledRef.current = false
    setSrc('')
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
          setSrc(
            `/api/v1/preview/sessions/${detail.session_hash}/files/${detail.filename}`,
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
    return <p className="text-xs text-red-500">{error}</p>
  }
  if (src === '') {
    return <p className="text-xs text-sea-ink-soft/50">正在加载预览…</p>
  }
  return <PreviewFrame src={src} className="h-80 w-full" />
}
