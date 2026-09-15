import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Copy, GitBranch, GripVertical, Settings2 } from 'lucide-react'
import { toast } from 'sonner'
import { PreviewFileViewer } from '#/components/preview/file-viewer'
import { previewKindFromFilename } from '#/lib/preview-file'
import { useForkPage, useSwitchActiveVersion } from '#/hooks/usePages'
import type { PageFileItem, PageItem, PageVersionItem } from '#/lib/models/response/pages'

function isRenderable(filename: string) {
  return /\.(html|htm|md)$/i.test(filename)
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
  const [pos, setPos] = useState({ x: 24, y: 24 })
  const drag = useRef<{ ox: number; oy: number; moved: boolean } | null>(null)
  const pillRef = useRef<HTMLButtonElement | null>(null)
  const fork = useForkPage()
  const switchActive = useSwitchActiveVersion()

  const activeFile = resolveActiveFile(filepath, version.entry_filename, files)
  // lumina_frame=1 标记 iframe 场景，后端据此直出文件，规避 Accept 启发式误判
  const src = activeFile
    ? `/pages/${projectName}/${slug}/${encodeURIComponent(activeFile)}?v=${encodeURIComponent(version.created_at)}&lumina_frame=1`
    : ''
  const renderable = files.filter((file) => isRenderable(file.filename))
  const advancedFiles = files.filter((file) => !isRenderable(file.filename))
  const inspectorFile = useMemo(() => {
    if (!activeFile || isRenderable(activeFile)) return null
    return activeFile
  }, [activeFile])

  useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (event.data?.type !== 'lumina:navigate' || typeof event.data.href !== 'string') {
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

  const selectFile = (filename: string) => {
    navigate({
      to: '/pages/$projectName/$slug/$',
      params: { projectName, slug, _splat: filename },
      replace: true,
    })
    setOpen(false)
  }

  return (
    <div className="relative h-screen overflow-hidden bg-sand">
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

      <button
        ref={pillRef}
        type="button"
        className="absolute z-30 flex items-center gap-2 border border-line bg-foam/95 px-3 py-1.5 text-xs text-sea-ink shadow-sm"
        style={{ right: pos.x, top: pos.y }}
        onPointerDown={(event) => {
          drag.current = { ox: event.clientX, oy: event.clientY, moved: false }
        }}
        onPointerMove={(event) => {
          if (!drag.current) return
          const dx = event.clientX - drag.current.ox
          const dy = event.clientY - drag.current.oy
          if (Math.hypot(dx, dy) > 4) drag.current.moved = true
          if (drag.current.moved) {
            // x 为 right 偏移（向左拖增大）：钳制在视口内，防止胶囊被拖出后无法找回
            const { width, height } = pillRef.current?.getBoundingClientRect() ?? {
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
          if (drag.current && !drag.current.moved) setOpen(true)
          drag.current = null
        }}
      >
        <GripVertical className="size-3.5 text-sea-ink-soft" />
        ✦ {page.project_name} / {page.slug} {version.version}
      </button>

      {open ? (
        <aside className="absolute inset-y-0 right-0 z-40 flex w-[min(360px,92vw)] flex-col border-l border-line bg-foam">
          <div className="flex items-center justify-between border-b border-line px-4 py-3">
            <p className="text-sm font-semibold text-sea-ink">Pages</p>
            <button type="button" onClick={() => setOpen(false)} className="text-xs">
              关闭
            </button>
          </div>
          <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-4 text-sm">
            <section>
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                元信息
              </p>
              <p className="mt-2 font-medium text-sea-ink">{page.title}</p>
              <p className="text-xs text-sea-ink-soft">
                {page.access_mode === 'password' ? '密码保护' : '公开访问'} · {version.version}
              </p>
            </section>
            <section>
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                版本
              </p>
              <ul className="mt-2 space-y-1">
                {versions.map((item) => (
                  <li key={item.id} className="flex items-center justify-between gap-2">
                    <span className={item.is_active ? 'text-lagoon-deep' : 'text-sea-ink-soft'}>
                      {item.version}
                    </span>
                    {!item.is_active ? (
                      <button
                        type="button"
                        className="text-xs"
                        onClick={() =>
                          switchActive.mutate({ id: page.id, versionId: item.id })
                        }
                      >
                        设为生效
                      </button>
                    ) : (
                      <span className="text-[10px] text-lagoon-deep">当前</span>
                    )}
                  </li>
                ))}
              </ul>
            </section>
            <section>
              <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-lagoon-deep">
                可渲染页面
              </p>
              <ul className="mt-2 space-y-1">
                {renderable.map((file) => (
                  <li key={file.id}>
                    <button
                      type="button"
                      className={`w-full text-left ${file.filename === activeFile ? 'text-lagoon-deep' : 'text-sea-ink'}`}
                      onClick={() => selectFile(file.filename)}
                    >
                      {file.filename}
                    </button>
                  </li>
                ))}
              </ul>
            </section>
            <section>
              <button
                type="button"
                className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sea-ink-soft"
                onClick={() => setAdvanced((value) => !value)}
              >
                高级资源 {advanced ? '▾' : '▸'}
              </button>
              {advanced ? (
                <ul className="mt-2 space-y-1">
                  {advancedFiles.map((file) => (
                    <li key={file.id}>
                      <button
                        type="button"
                        className="w-full text-left text-sea-ink-soft"
                        onClick={() => selectFile(file.filename)}
                      >
                        {file.filename}
                      </button>
                    </li>
                  ))}
                </ul>
              ) : null}
            </section>
          </div>
          <div className="space-y-2 border-t border-line p-4">
            <button
              type="button"
              className="flex w-full items-center gap-2 text-left text-sm"
              onClick={async () => {
                await navigator.clipboard.writeText(window.location.href)
                toast.success('已复制当前路径')
              }}
            >
              <Copy className="size-3.5" /> 复制当前路径
            </button>
            <button
              type="button"
              className="flex w-full items-center gap-2 text-left text-sm"
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
              <GitBranch className="size-3.5" /> Fork 到新预览
            </button>
            <a
              href="/console/pages"
              className="flex items-center gap-2 text-sm text-sea-ink-soft"
            >
              <Settings2 className="size-3.5" /> 管理员端口 · 页面安全设置
            </a>
          </div>
        </aside>
      ) : null}
    </div>
  )
}
