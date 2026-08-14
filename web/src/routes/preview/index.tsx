import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useCallback, useEffect, useState } from 'react'
import { FileCode2, FolderOpen } from 'lucide-react'

import { PreviewFrame } from '#/components/interact/primitives/preview-frame'
import { usePreviewWebSocket } from '#/hooks/usePreviewWebSocket'
import type {
  PreviewFileItem,
  PreviewSessionItem,
} from '#/lib/models/response/preview'

interface PreviewSearch {
  session?: string
  file?: string
}

/** preview_sync 消息的 data 结构（{ session, files }，字段 snake_case） */
interface PreviewSyncData {
  session: PreviewSessionItem
  files: PreviewFileItem[]
}

export const Route = createFileRoute('/preview/')({
  validateSearch: (search: Record<string, unknown>): PreviewSearch => {
    return {
      session: typeof search.session === 'string' ? search.session : undefined,
      file: typeof search.file === 'string' ? search.file : undefined,
    }
  },
  component: PreviewPage,
})

function PreviewPage() {
  const search = useSearch({ from: '/preview/' })
  const navigate = useNavigate()

  const [files, setFiles] = useState<PreviewFileItem[]>([])
  const [activeFile, setActiveFile] = useState('')
  const [sessionTitle, setSessionTitle] = useState('')
  const [error, setError] = useState('')

  const hash = search.session

  // WS 实时同步：连接快照 / 文件变更均通过 preview_sync 消息驱动，替代单次 REST 拉取
  const handleSync = useCallback(
    (data: any) => {
      if (!data?.session) return
      const syncData = data as PreviewSyncData
      setSessionTitle(syncData.session.title)
      setFiles(syncData.files)

      // 文件变更时若当前激活文件仍存在则保留，否则按 深链参数 → 首个 HTML → 首个文件 回退
      let next = ''
      if (activeFile && syncData.files.some((f) => f.filename === activeFile)) {
        next = activeFile
      } else if (
        search.file &&
        syncData.files.some((f) => f.filename === search.file)
      ) {
        next = search.file
      } else {
        const htmlFile = syncData.files.find(
          (f) => f.filename.endsWith('.html') || f.filename.endsWith('.htm'),
        )
        next = htmlFile ? htmlFile.filename : (syncData.files[0]?.filename ?? '')
      }

      if (next !== activeFile) {
        setActiveFile(next)
        // 同步 URL 深链（首次选择或文件回退时）
        if (next && next !== search.file) {
          navigate({
            to: '/preview',
            search: { session: hash, file: next },
            replace: true,
          })
        }
      }
    },
    [activeFile, search.file, hash, navigate],
  )

  const { status } = usePreviewWebSocket(hash ?? null, {
    onSync: handleSync,
  })

  // WS 状态驱动错误态：rejected 表示会话不存在，连接中/已连接时清除错误
  useEffect(() => {
    if (status === 'rejected') {
      setError('预览会话不存在')
    } else if (status === 'connecting' || status === 'connected') {
      setError('')
    }
  }, [status])

  // 加载态由 WS 状态驱动：idle / connecting 视为加载中
  const isLoading = status === 'idle' || status === 'connecting'

  const selectFile = (filename: string) => {
    setActiveFile(filename)
    navigate({
      to: '/preview',
      search: { session: hash, file: filename },
      replace: true,
    })
  }

  if (!hash) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-sea-ink-soft">
          缺少预览会话参数（?session=hash）
        </p>
      </div>
    )
  }

  // iframe src 追加 cache-buster（当前激活文件的 updated_at），文件变更后强制刷新预览
  const activeUpdatedAt =
    files.find((f) => f.filename === activeFile)?.updated_at ?? ''
  const src = activeFile
    ? `/api/v1/preview/sessions/${hash}/files/${activeFile}?v=${encodeURIComponent(activeUpdatedAt)}`
    : ''

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      {/* 左侧文件列表 */}
      <aside className="flex w-60 shrink-0 flex-col border-r border-line bg-surface/50">
        <div className="border-b border-line px-4 py-3">
          <h2 className="truncate text-sm font-semibold text-sea-ink">
            {sessionTitle || '预览工作区'}
          </h2>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {isLoading ? (
            <p className="px-2 py-4 text-xs text-sea-ink-soft/50">加载中…</p>
          ) : files.length === 0 ? (
            <p className="px-2 py-4 text-xs text-sea-ink-soft/50">暂无文件</p>
          ) : (
            <ul className="space-y-0.5">
              {files.map((f) => (
                <li key={f.id}>
                  <button
                    type="button"
                    onClick={() => selectFile(f.filename)}
                    className={`flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-left text-xs transition-colors ${
                      f.filename === activeFile
                        ? 'bg-lagoon/10 text-lagoon-deep'
                        : 'text-sea-ink-soft hover:bg-line/30'
                    }`}
                  >
                    <FileCode2 className="size-3.5 shrink-0" aria-hidden />
                    <span className="truncate">{f.filename}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>

      {/* 右侧预览区 */}
      <main className="flex min-w-0 flex-1 flex-col">
        {error ? (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-red-500">{error}</p>
          </div>
        ) : isLoading ? (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-sea-ink-soft/50">加载中…</p>
          </div>
        ) : src ? (
          <PreviewFrame src={src} className="flex-1" />
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center gap-2">
            <FolderOpen className="size-6 text-sea-ink-soft/40" aria-hidden />
            <p className="text-sm text-sea-ink-soft/50">选择左侧文件预览</p>
          </div>
        )}
      </main>
    </div>
  )
}
