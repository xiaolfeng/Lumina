import { Rocket, Sparkles } from 'lucide-react'
import type { PreviewHeaderControls } from '#/hooks/usePreviewHeader'
import { usePreviewHeader } from '#/hooks/usePreviewHeader'
import type { WorkbenchDevice } from '#/components/preview/workbench-canvas'

type ViewOptionId = WorkbenchDevice | 'source'

const VIEW_OPTIONS: { id: ViewOptionId; label: string }[] = [
  { id: 'desktop', label: '桌面' },
  { id: 'tablet', label: '平板' },
  { id: 'mobile', label: '手机' },
  { id: 'source', label: '源码' },
]

export function PreviewBrandHeader({
  title,
  controls: propControls,
}: {
  title: string | null
  controls?: PreviewHeaderControls | null
}) {
  const context = usePreviewHeader()
  const controls = propControls !== undefined ? propControls : context.controls

  const activeView: ViewOptionId = controls?.sourceMode
    ? 'source'
    : (controls?.device ?? 'desktop')

  const handleSelectView = (id: ViewOptionId) => {
    if (!controls) return
    if (id === 'source') {
      controls.setSourceMode(true)
      controls.setDevice('desktop')
    } else {
      controls.setSourceMode(false)
      controls.setDevice(id)
    }
  }

  return (
    <header className="flex h-12 shrink-0 items-center justify-between gap-3 border-b border-line bg-header-bg px-4">
      {/* 左侧：品牌 Logo 与 Preview 标识 */}
      <div className="flex shrink-0 items-center gap-2">
        <Sparkles className="size-4 text-lagoon" aria-hidden />
        <span className="display-title text-sm font-bold tracking-tight text-sea-ink">
          Lumina
        </span>
        <span className="hidden items-center gap-1.5 sm:inline-flex">
          <span className="h-px w-3 bg-lagoon/40" />
          <span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-lagoon-deep">
            Preview
          </span>
        </span>
      </div>

      {/* 居中：桌面、平板、手机、源码切换 ButtonGroup（四个合一） */}
      {controls ? (
        <div className="flex flex-1 justify-center px-2">
          <div
            role="group"
            aria-label="视口与源码模式切换"
            className="flex items-center gap-px border border-line bg-chip-bg p-0.5"
          >
            {VIEW_OPTIONS.map((item) => {
              const isActive = activeView === item.id
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => handleSelectView(item.id)}
                  className={`cursor-pointer px-2.5 py-0.5 text-xs transition-colors ${
                    isActive
                      ? 'bg-foam font-semibold text-sea-ink shadow-xs'
                      : 'text-sea-ink-soft hover:text-sea-ink'
                  }`}
                >
                  {item.label}
                </button>
              )
            })}
          </div>
        </div>
      ) : (
        <div className="flex-1" />
      )}

      {/* 右侧：名字标题靠右侧 + 晋升为 Pages 按钮 */}
      <div className="flex shrink-0 items-center gap-2.5">
        {controls?.sourcePageSlug ? (
          <span className="hidden border border-green-600/30 bg-green-600/10 px-1.5 py-0.5 text-[10px] font-semibold text-green-700 lg:inline-flex">
            迭代 /{controls.sourcePageSlug}
          </span>
        ) : null}

        {title ? (
          <h1
            className="min-w-0 max-w-[min(30vw,16rem)] truncate text-right text-sm font-medium tracking-tight text-sea-ink"
            title={title}
          >
            {title}
          </h1>
        ) : null}

        {controls?.onPromoteClick ? (
          <button
            type="button"
            onClick={controls.onPromoteClick}
            className="inline-flex cursor-pointer items-center gap-1.5 bg-lagoon px-2.5 py-1 text-xs font-semibold text-white shadow-xs transition-colors hover:bg-lagoon-deep"
          >
            <Rocket className="size-3.5" aria-hidden />
            <span>晋升为 Pages</span>
          </button>
        ) : null}
      </div>
    </header>
  )
}
