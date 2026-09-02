import { Sparkles } from 'lucide-react'

export function PreviewBrandHeader({ title }: { title: string | null }) {
  return (
    <header className="flex h-12 shrink-0 items-center gap-2 border-b border-line bg-header-bg px-4">
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

      <div className="flex-1" />

      {title && (
        <h1
          className="min-w-0 max-w-[min(50%,20rem)] truncate text-sm font-medium tracking-tight text-sea-ink"
          title={title}
        >
          {title}
        </h1>
      )}
    </header>
  )
}
