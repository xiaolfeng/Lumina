import { Link } from '@tanstack/react-router'
import type { NavRef } from '#/lib/api-client'

interface BreadcrumbProps {
  items?: NavRef[]
  wikiId: string
}

export function Breadcrumb({ items, wikiId }: BreadcrumbProps) {
  if (!items || items.length === 0) {
    return null
  }

  return (
    <nav aria-label="breadcrumb">
      <ol className="flex flex-wrap items-center gap-1.5 text-sm">
        {items.map((item, index) => {
          const isLast = index === items.length - 1

          return (
            <li key={`${item.path}-${index}`} className="flex items-center gap-1.5">
              {index > 0 && (
                <span className="text-sea-ink-soft/40 select-none">&gt;</span>
              )}
              {isLast ? (
                <span className="text-sea-ink-soft font-medium" aria-current="page">
                  {item.title}
                </span>
              ) : (
                <Link
                  to="/wiki/$wikiId/$"
                  params={{ wikiId, _splat: item.path }}
                  className="text-sea-ink-soft hover:text-lagoon transition-colors"
                >
                  {item.title}
                </Link>
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
