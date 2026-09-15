import { useNavigate } from '@tanstack/react-router'
import { useCallback, useEffect, useRef, useState } from 'react'
import { FileCode2, FolderOpen, Rocket } from 'lucide-react'

import {
  DEVICE_PRESETS,
  WorkbenchCanvas,
  type WorkbenchDevice,
} from '#/components/preview/workbench-canvas'
import { PromoteDialog } from '#/components/preview/promote-dialog'
import { usePreviewHeader } from '#/hooks/usePreviewHeader'
import { usePreviewWebSocket } from '#/hooks/usePreviewWebSocket'
import { getPreviewSessionDetail } from '#/lib/apis/preview'
import type {
  PreviewFileItem,
  PreviewSessionItem,
} from '#/lib/models/response/preview'

interface PreviewSyncData {
  session: PreviewSessionItem
  files: PreviewFileItem[]
}

export function PreviewWorkbenchPage({
  sessionHash,
  requestedFile,
}: {
  sessionHash: string
  requestedFile: string
}) {
  const navigate = useNavigate()
  const { setTitle } = usePreviewHeader()
  const [files, setFiles] = useState<PreviewFileItem[]>([])
  const [session, setSession] = useState<PreviewSessionItem | null>(null)
  const [error, setError] = useState('')
  const [sourceMode, setSourceMode] = useState(false)
  const [device, setDevice] = useState<WorkbenchDevice>('desktop')
  const [promoteOpen, setPromoteOpen] = useState(false)
  const [syncSeq, setSyncSeq] = useState(0)
  const [activeFile, setActiveFile] = useState(requestedFile)

  const handleSync = useCallback(
    (data: any) => {
      if (!data?.session || data.session.status !== 'active') {
        if (data?.event_type === 'delete_session' || data?.session) {
          setFiles([])
          setSession(null)
          setError('预览会话已被删除')
        }
        return
      }
      const syncData = data as PreviewSyncData
      setSession(syncData.session)
      setFiles(syncData.files)
      setSyncSeq((seq) => seq + 1)
      let next = ''
      if (activeFile && syncData.files.some((file) => file.filename === activeFile)) {
        next = activeFile
      } else if (
        requestedFile &&
        syncData.files.some((file) => file.filename === requestedFile)
      ) {
        next = requestedFile
      } else {
        const htmlFile = syncData.files.find(
          (file) =>
            file.filename.endsWith('.html') || file.filename.endsWith('.htm'),
        )
        next = htmlFile ? htmlFile.filename : (syncData.files[0]?.filename ?? '')
      }
      if (next && next !== activeFile) {
        setActiveFile(next)
        if (next !== requestedFile) {
          navigate({
            to: '/preview/$sessionHash/$',
            params: { sessionHash, _splat: next },
            replace: true,
          })
        }
      }
    },
    [activeFile, requestedFile, navigate, sessionHash],
  )

  const { status } = usePreviewWebSocket(sessionHash, { onSync: handleSync })

  const handleSyncRef = useRef(handleSync)
  handleSyncRef.current = handleSync
  useEffect(() => {
    let cancelled = false
    void getPreviewSessionDetail(sessionHash)
      .then((res) => {
        if (!cancelled && res.data?.session) handleSyncRef.current(res.data)
      })
      .catch(() => undefined)
    return () => {
      cancelled = true
    }
  }, [sessionHash])

  useEffect(() => {
    if (status === 'rejected') {
      setError('预览会话不存在')
    } else if (status === 'connecting' || status === 'connected') {
      setError('')
    }
  }, [status])

  useEffect(() => {
    setTitle(session?.title ?? null)
    return () => setTitle(null)
  }, [session?.title, setTitle])

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (event.data?.type !== 'lumina:navigate' || typeof event.data.href !== 'string') {
        return
      }
      const href = event.data.href.split(/[?#]/)[0].split('/').pop()
      if (!href) return
      setActiveFile(href)
      navigate({
        to: '/preview/$sessionHash/$',
        params: { sessionHash, _splat: href },
        replace: true,
      })
    }
    window.addEventListener('message', onMessage)
    return () => window.removeEventListener('message', onMessage)
  }, [navigate, sessionHash])

  const selectFile = (filename: string) => {
    setActiveFile(filename)
    navigate({
      to: '/preview/$sessionHash/$',
      params: { sessionHash, _splat: filename },
      replace: true,
    })
  }

  const isLoading =
    session === null && (status === 'idle' || status === 'connecting')
  const activeUpdatedAt =
    files.find((file) => file.filename === activeFile)?.updated_at ?? ''
  const src = activeFile
    ? `/preview/${sessionHash}/${encodeURIComponent(activeFile)}?v=${encodeURIComponent(
        activeUpdatedAt,
      )}&_lumina_sync=${syncSeq}&lumina_frame=1`
    : ''
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
      <div className="flex shrink-0 flex-wrap items-center gap-2 overflow-x-auto border-b border-line px-3 py-2">
        <span className="border border-lagoon/40 bg-[#fdf5ea] px-1.5 py-0.5 text-[10px] font-bold tracking-wide text-lagoon-deep">
          草稿
        </span>
        {session?.source_page_id ? (
          <span className="border border-green-600/30 bg-green-600/10 px-1.5 py-0.5 text-[10px] font-semibold text-green-700">
            迭代{session.source_page_slug ? ` /${session.source_page_slug}` : ''}
          </span>
        ) : null}
        {files.length > 0 ? (
          <select
            aria-label="切换预览文件"
            className="max-w-40 border border-line bg-transparent px-1.5 py-0.5 text-xs md:hidden"
            value={activeFile}
            onChange={(event) => selectFile(event.target.value)}
          >
            {files.map((file) => (
              <option key={file.id} value={file.filename}>
                {file.filename}
              </option>
            ))}
          </select>
        ) : null}
        {!sourceMode ? (
          <div className="flex items-center gap-px border border-line bg-chip-bg p-0.5">
            {DEVICE_PRESETS.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => setDevice(item.id)}
                className={`px-2 py-0.5 text-[11px] ${
                  device === item.id
                    ? 'bg-foam font-semibold text-sea-ink'
                    : 'text-sea-ink-soft hover:text-sea-ink'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        ) : null}
        <div className="flex-1" />
        <div className="flex border border-line bg-chip-bg p-0.5">
          <button
            type="button"
            className={`px-2 py-0.5 text-[11px] ${
              !sourceMode ? 'bg-foam font-semibold text-sea-ink' : 'text-sea-ink-soft'
            }`}
            onClick={() => setSourceMode(false)}
          >
            视口
          </button>
          <button
            type="button"
            className={`px-2 py-0.5 text-[11px] ${
              sourceMode ? 'bg-foam font-semibold text-sea-ink' : 'text-sea-ink-soft'
            }`}
            onClick={() => {
              setSourceMode(true)
              setDevice('desktop')
            }}
          >
            源码
          </button>
        </div>
        <button
          type="button"
          className="inline-flex items-center gap-1 bg-lagoon px-2.5 py-1 text-xs font-semibold text-white"
          onClick={() => setPromoteOpen(true)}
        >
          <Rocket className="size-3.5" />
          晋升为 Pages
        </button>
      </div>
      <div className="flex min-h-0 flex-1 overflow-hidden">
        <aside className="hidden w-60 shrink-0 flex-col border-r border-line bg-surface/50 md:flex">
          <div className="border-b border-line px-4 py-3">
            <p className="text-[10px] font-semibold uppercase tracking-[0.15em] text-lagoon-deep">
              文件
            </p>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            {isLoading ? (
              <p className="px-2 py-4 text-xs text-sea-ink-soft/50">加载中…</p>
            ) : files.length === 0 ? (
              <p className="px-2 py-4 text-xs text-sea-ink-soft/50">暂无文件</p>
            ) : (
              <ul className="space-y-0.5">
                {files.map((file) => (
                  <li key={file.id}>
                    <button
                      type="button"
                      onClick={() => selectFile(file.filename)}
                      className={`flex w-full items-center gap-2 px-2.5 py-1.5 text-left text-xs ${
                        file.filename === activeFile
                          ? 'bg-lagoon/10 text-lagoon-deep'
                          : 'text-sea-ink-soft hover:bg-line/30'
                      }`}
                    >
                      <FileCode2 className="size-3.5 shrink-0" aria-hidden />
                      <span className="truncate">{file.filename}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </aside>
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
            <WorkbenchCanvas
              src={src}
              filename={activeFile}
              sourceMode={sourceMode}
              device={device}
            />
          ) : (
            <div className="flex flex-1 flex-col items-center justify-center gap-2">
              <FolderOpen className="size-6 text-sea-ink-soft/40" aria-hidden />
              <p className="text-sm text-sea-ink-soft/50">选择左侧文件预览</p>
            </div>
          )}
        </main>
      </div>
      <PromoteDialog
        open={promoteOpen}
        onOpenChange={setPromoteOpen}
        sessionId={session?.id}
        defaultSlug={session?.source_page_slug}
        defaultTitle={session?.title}
        forked={Boolean(session?.source_page_id)}
      />
    </div>
  )
}
