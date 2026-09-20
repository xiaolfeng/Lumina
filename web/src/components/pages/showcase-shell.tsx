import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Copy, GitBranch, GripVertical, Settings2 } from 'lucide-react'
import { toast } from 'sonner'
import { PagesBrandHeader } from '#/components/pages/brand-header'
import { PreviewFileViewer } from '#/components/preview/file-viewer'
import { previewKindFromFilename } from '#/lib/preview-file'
import { useForkPage, useSwitchActiveVersion } from '#/hooks/usePages'
import type {
  PageFileItem,
  PageItem,
  PageVersionItem,
} from '#/lib/models/response/pages'

export function isRenderable(filename: string) {
  return /\.(html|htm|md|lpw)$/i.test(filename)
}

/** 解析当前激活文件：URL 中的文件名必须存在于版本文件清单，否则回退版本入口，再回退首个文件 */
export function resolveActiveFile(
  filepath: string,
  entryFilename: string,
  files: { filename: string }[],
): string {
  const names = new Set(files.map((file) => file.filename))
  if (filepath && names.has(filepath)) return filepath
  if (entryFilename && names.has(entryFilename)) return entryFilename
  return files[0]?.filename ?? ''
}

/**
 * 拼接展示态 iframe src。
 * v 是后端语义化版本号（如 v1.0.0），生效版本渲染必须省略 v（空标签走 LatestVersionID）；
 * 缓存戳用独立的 _t 承载，禁止把 created_at 时间戳塞进 v，否则后端按版本号查库必然 404。
 * lumina_frame=1 标记 iframe 场景，后端据此直出文件，规避 Accept 启发式误判。
 */
export function buildShowcaseSrc(
  projectName: string,
  slug: string,
  filename: string,
  cacheKey: string,
): string {
  if (!filename) return ''
  return `/pages/${projectName}/${slug}/${encodeURIComponent(filename)}?_t=${encodeURIComponent(cacheKey)}&lumina_frame=1`
}

export function ShowcaseShell({
  projectName,
  slug,
  filepath,
  page,
  version,
  files,
  versions,
}: {
  projectName: string
  slug: string
  filepath: string
  page: PageItem
  version: PageVersionItem
  files: PageFileItem[]
  versions: PageVersionItem[]
}) {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [advanced, setAdvanced] = useState(false)
  const [pos, setPos] = useState({ x: 20, y: 16 })
  const drag = useRef<{ ox: number; oy: number } | null>(null)
  const isDraggingRef = useRef(false)
  const dragMovedRef = useRef(false)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const fork = useForkPage()
  const switchActive = useSwitchActiveVersion()

  const activeFile = resolveActiveFile(filepath, version.entry_filename, files)
  const src = buildShowcaseSrc(
    projectName,
    slug,
    activeFile,
    version.created_at,
  )
  const renderable = files.filter((file) => isRenderable(file.filename))
  const advancedFiles = files.filter((file) => !isRenderable(file.filename))
  const inspectorFile = useMemo(() => {
    if (!activeFile || isRenderable(activeFile)) return null
    return activeFile
  }, [activeFile])

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (
        event.data?.type !== 'lumina:navigate' ||
        typeof event.data.href !== 'string'
      ) {
        return
      }
      const href = event.data.href.split(/[?#]/)[0].split('/').pop()
      if (!href) return
      navigate({
        to: '/pages/$projectName/$slug/$',
        params: { projectName, slug, _splat: href },
        replace: true,
      })
    }
    window.addEventListener('message', onMessage)
    return () => window.removeEventListener('message', onMessage)
  }, [navigate, projectName, slug])

  useEffect(() => {
    if (!open) return
    const handlePointerDownOutside = (event: PointerEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setOpen(false)
      }
    }
    window.addEventListener('pointerdown', handlePointerDownOutside)
    return () => {
      window.removeEventListener('pointerdown', handlePointerDownOutside)
    }
  }, [open])

  const selectFile = (filename: string) => {
    navigate({
      to: '/pages/$projectName/$slug/$',
      params: { projectName, slug, _splat: filename },
      replace: true,
    })
    setOpen(false)
  }

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-sand">
      {/* 顶部 Header：遵循设计稿与 Lumina 规范 */}
      <PagesBrandHeader
        projectName={projectName}
        slug={slug}
        title={page.title}
      />

      {/* 主画布与悬浮动 Bar 容器 */}
      <main className="relative min-h-0 flex-1 overflow-hidden">
        {inspectorFile ? (
          <PreviewFileViewer
            kind={previewKindFromFilename(inspectorFile)}
            src={`/pages/${projectName}/${slug}/${encodeURIComponent(inspectorFile)}`}
            filename={inspectorFile}
          />
        ) : src ? (
          <PreviewFileViewer
            kind={previewKindFromFilename(activeFile)}
            src={src}
            filename={activeFile}
          />
        ) : null}

        {/* 悬浮动 Bar（点击展开悬浮卡片面板，随 Bar 悬浮而不是全屏右侧展开） */}
        <div
          ref={containerRef}
          className="absolute z-30 flex flex-col border border-line bg-foam/95 shadow-md backdrop-blur-xs transition-shadow duration-150"
          style={{ right: pos.x, top: pos.y }}
        >
          {/* 胶囊 Bar 头部：支持拖拽，点击展开/折叠 */}
          <div
            role="button"
            tabIndex={0}
            aria-label="展开或折叠页面管理卡片"
            className="flex cursor-grab select-none items-center gap-2 px-3 py-1.5 text-xs text-sea-ink transition-colors hover:bg-sand/60 active:cursor-grabbing"
            onPointerDown={(event) => {
              drag.current = { ox: event.clientX, oy: event.clientY }
              isDraggingRef.current = true
              dragMovedRef.current = false
            }}
            onPointerMove={(event) => {
              if (!isDraggingRef.current || !drag.current) return
              const dx = event.clientX - drag.current.ox
              const dy = event.clientY - drag.current.oy
              if (Math.hypot(dx, dy) > 4) {
                dragMovedRef.current = true
              }
              if (dragMovedRef.current) {
                const { width, height } =
                  containerRef.current?.getBoundingClientRect() ?? {
                    width: 0,
                    height: 0,
                  }
                const maxX = window.innerWidth - width - 8
                const maxY = window.innerHeight - height - 8
                setPos((prev) => ({
                  x: Math.max(8, Math.min(prev.x - dx, maxX)),
                  y: Math.max(8, Math.min(prev.y + dy, maxY)),
                }))
                drag.current.ox = event.clientX
                drag.current.oy = event.clientY
              }
            }}
            onPointerUp={() => {
              isDraggingRef.current = false
              setTimeout(() => {
                dragMovedRef.current = false
              }, 100)
            }}
            onClick={() => {
              if (dragMovedRef.current) return
              setOpen((prev) => !prev)
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                setOpen((prev) => !prev)
              }
            }}
          >
            <GripVertical className="size-3.5 shrink-0 text-sea-ink-soft" />
            <span className="font-bold text-lagoon">✦</span>
            <span className="font-mono font-semibold">
              {page.project_name} / {page.slug}
            </span>
            <span className="bg-sand-tint px-1.5 py-0.5 text-[10px] text-sea-ink-soft">
              {version.version}
            </span>
            <span className="ml-0.5 text-[10px] text-sea-ink-soft transition-transform duration-200">
              {open ? '▴' : '▾'}
            </span>
          </div>

          {/* 展开悬浮面板卡片（挂在 Bar 容器内部，跟随 Bar 悬浮展开） */}
          {open ? (
            <div className="max-h-[min(540px,calc(100vh-130px))] w-[min(340px,calc(100vw-32px))] space-y-4 overflow-y-auto border-t border-line bg-foam p-3.5 text-xs animate-in fade-in zoom-in-95 duration-150">
              {/* 元信息区 */}
              <section>
                <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                  元信息
                </p>
                <p className="mt-1 text-sm font-medium text-sea-ink">
                  {page.title}
                </p>
                <p className="mt-0.5 text-[11px] text-sea-ink-soft">
                  {page.access_mode === 'password'
                    ? '🔒 密码保护'
                    : '🌐 公开访问'}{' '}
                  · 不可变快照
                </p>
              </section>

              {/* 生效指针与版本历史 */}
              <section>
                <div className="flex items-center justify-between">
                  <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                    版本
                  </p>
                  <span className="font-mono text-[10px] font-bold text-emerald-700">
                    当前生效: {version.version}
                  </span>
                </div>
                <ul className="mt-1.5 space-y-1">
                  {versions.map((item) => (
                    <li
                      key={item.id}
                      className="flex items-center justify-between gap-2 py-0.5"
                    >
                      <span
                        className={
                          item.is_active
                            ? 'font-semibold text-lagoon-deep'
                            : 'text-sea-ink-soft'
                        }
                      >
                        {item.version}
                      </span>
                      {!item.is_active ? (
                        <button
                          type="button"
                          className="cursor-pointer text-[11px] text-sea-ink-soft underline hover:text-sea-ink"
                          onClick={() =>
                            switchActive.mutate({
                              id: page.id,
                              versionId: item.id,
                            })
                          }
                        >
                          设为生效
                        </button>
                      ) : (
                        <span className="text-[10px] text-emerald-700">
                          当前
                        </span>
                      )}
                    </li>
                  ))}
                </ul>
              </section>

              {/* 可渲染页面直切区 */}
              <section>
                <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                  可渲染页面
                </p>
                <ul className="mt-1.5 space-y-1">
                  {renderable.map((file) => {
                    const isSelected = file.filename === activeFile
                    return (
                      <li key={file.id}>
                        <button
                          type="button"
                          className={`block w-full cursor-pointer truncate border px-2 py-1.5 text-left transition-colors ${
                            isSelected
                              ? 'border-lagoon bg-sand-tint font-semibold text-lagoon-deep'
                              : 'border-line/60 text-sea-ink hover:border-line hover:bg-sand/60'
                          }`}
                          onClick={() => selectFile(file.filename)}
                        >
                          {file.filename}
                        </button>
                      </li>
                    )
                  })}
                </ul>
              </section>

              {/* 高级资源折叠区 */}
              {advancedFiles.length > 0 ? (
                <section>
                  <button
                    type="button"
                    className="flex cursor-pointer items-center gap-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-sea-ink-soft hover:text-sea-ink"
                    onClick={() => setAdvanced((value) => !value)}
                  >
                    <span>高级资源</span>
                    <span>{advanced ? '▾' : '▸'}</span>
                  </button>
                  {advanced ? (
                    <ul className="mt-1.5 space-y-1">
                      {advancedFiles.map((file) => (
                        <li key={file.id}>
                          <button
                            type="button"
                            className="block w-full cursor-pointer truncate font-mono text-[11px] text-sea-ink-soft hover:bg-sand/60 hover:text-sea-ink"
                            onClick={() => selectFile(file.filename)}
                          >
                            {file.filename}
                          </button>
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </section>
              ) : null}

              {/* 底部快捷操作条 */}
              <div className="space-y-1.5 border-t border-line pt-3">
                <button
                  type="button"
                  className="flex w-full cursor-pointer items-center gap-2 text-left text-xs text-sea-ink hover:text-lagoon-deep"
                  onClick={async () => {
                    await navigator.clipboard.writeText(window.location.href)
                    toast.success('已复制当前路径')
                  }}
                >
                  <Copy className="size-3.5 text-sea-ink-soft" /> 复制当前路径
                </button>
                <button
                  type="button"
                  className="flex w-full cursor-pointer items-center gap-2 text-left text-xs text-sea-ink hover:text-lagoon-deep"
                  onClick={() =>
                    fork.mutate(
                      { id: page.id, versionId: version.id },
                      {
                        onSuccess: (res) => {
                          const url = res.data?.preview_url
                          if (url) window.location.href = url
                        },
                      },
                    )
                  }
                >
                  <GitBranch className="size-3.5 text-sea-ink-soft" /> Fork
                  到新预览
                </button>
                <a
                  href="/console/pages"
                  className="flex items-center gap-2 text-xs text-sea-ink-soft hover:text-sea-ink"
                >
                  <Settings2 className="size-3.5" /> 管理员端口 · 页面安全设置
                </a>
              </div>
            </div>
          ) : null}
        </div>
      </main>
    </div>
  )
}
