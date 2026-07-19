import { Link } from '@tanstack/react-router'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { NavRef } from '#/lib/api-client'

interface PrevNextProps {
  prev?: NavRef | null
  next?: NavRef | null
  wikiId: string
}

export function PrevNext({ prev, next, wikiId }: PrevNextProps) {
  if (!prev && !next) {
    return null
  }

  return (
    <nav
      aria-label="page navigation"
      className="flex items-stretch gap-4 border-t border-line/60 pt-6"
    >
      {/* Previous page */}
      <div className="flex flex-1">
        {prev && (
          <Link
            to="/wiki/$wikiId/$"
            params={{ wikiId, _splat: prev.path }}
            className="group flex flex-col gap-1 rounded-lg p-3 transition-colors hover:bg-accent/50 flex-1"
          >
            <span className="flex items-center gap-1 text-xs text-sea-ink-soft">
              <ChevronLeft className="size-3.5" />
              <span>上一页</span>
            </span>
            <span className="text-sm font-medium text-sea-ink-soft group-hover:text-lagoon transition-colors line-clamp-1">
              {prev.title}
            </span>
          </Link>
        )}
      </div>

      {/* Next page */}
      <div className="flex flex-1 justify-end">
        {next && (
          <Link
            to="/wiki/$wikiId/$"
            params={{ wikiId, _splat: next.path }}
            className="group flex flex-col gap-1 rounded-lg p-3 transition-colors hover:bg-accent/50 flex-1 items-end text-right"
          >
            <span className="flex items-center gap-1 text-xs text-sea-ink-soft">
              <span>下一页</span>
              <ChevronRight className="size-3.5" />
            </span>
            <span className="text-sm font-medium text-sea-ink-soft group-hover:text-lagoon transition-colors line-clamp-1">
              {next.title}
            </span>
          </Link>
        )}
      </div>
    </nav>
  )
}
