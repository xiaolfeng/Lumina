import { Sparkles } from 'lucide-react'

export function PagesBrandHeader({
  projectName,
  slug,
  title,
}: {
  projectName?: string
  slug?: string
  title?: string | null
}) {
  const displayTitle =
    title && title.trim() !== ''
      ? title
      : projectName && slug
        ? `${projectName} / ${slug}`
        : null

  return (
    <header className="flex h-12 shrink-0 items-center justify-between gap-3 border-b border-line bg-header-bg px-4">
      {/* 左侧：品牌 Logo 与 Pages 标识 */}
      <div className="flex shrink-0 items-center gap-2">
        <Sparkles className="size-4 text-lagoon" aria-hidden />
        <span className="display-title text-sm font-bold tracking-tight text-sea-ink">
          Lumina
        </span>
        <span className="hidden items-center gap-1.5 sm:inline-flex">
          <span className="h-px w-3 bg-lagoon/40" />
          <span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-lagoon-deep">
            Pages
          </span>
        </span>
      </div>

      {/* 右侧：页面标题 / 标识 */}
      {displayTitle ? (
        <h1
          className="min-w-0 max-w-[min(50%,24rem)] truncate text-right text-sm font-medium tracking-tight text-sea-ink"
          title={displayTitle}
        >
          {displayTitle}
        </h1>
      ) : null}
    </header>
  )
}
