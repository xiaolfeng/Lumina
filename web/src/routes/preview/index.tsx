import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { FileCode2, FolderOpen } from 'lucide-react'

import { PreviewFrame } from '#/components/interact/primitives/preview-frame'
import { getPreviewSessionByHash } from '#/lib/apis/preview'
import type { PreviewFileItem } from '#/lib/models/response/preview'

interface PreviewSearch {
  session?: string
  file?: string
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
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const hash = search.session

  // 首次加载会话：拉取文件列表并确定激活文件（URL file 参数优先，否则首个 HTML 文件）
  useEffect(() => {
    if (!hash) return
    setIsLoading(true)
    setError('')
    ;(async () => {
      try {
        const res = await getPreviewSessionByHash(hash)
        const data = res.data
        if (!data) {
          setError('预览会话不存在')
          return
        }
        setSessionTitle(data.session.title)
        setFiles(data.files)

        const htmlFile = data.files.find(
          (f) => f.filename.endsWith('.html') || f.filename.endsWith('.htm'),
        )
        let initial = ''
        if (search.file && data.files.some((f) => f.filename === search.file)) {
          initial = search.file
        } else if (htmlFile) {
          initial = htmlFile.filename
        } else if (data.files.length > 0) {
          initial = data.files[0].filename
        }
        setActiveFile(initial)

        // 同步 URL 深链
        if (initial && initial !== search.file) {
          navigate({
            to: '/preview',
            search: { session: hash, file: initial },
            replace: true,
          })
        }
      } catch {
        setError('加载预览会话失败')
      } finally {
        setIsLoading(false)
      }
    })()
  }, [hash])

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

  const src = activeFile
    ? `/api/v1/preview/sessions/${hash}/files/${activeFile}`
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
