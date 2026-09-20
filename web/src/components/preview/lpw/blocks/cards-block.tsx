import type React from 'react'
import type { LpwBlockSlotProps, LpwCardsProps } from '../types'

function sanitizeHref(href?: string): string | undefined {
  if (!href) return undefined
  if (/^(javascript|data|vbscript):/i.test(href.trim())) {
    return undefined
  }
  return href
}

export const CardsBlock: React.FC<LpwBlockSlotProps<LpwCardsProps>> = ({
  props,
}) => {
  return (
    <div className="my-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {props.items.map((item, idx) => {
        const safeHref = sanitizeHref(item.href)
        const content = (
          <>
            <div className="font-semibold text-sm text-sea-ink">
              {item.title}
            </div>
            {item.description && (
              <p className="mt-1.5 text-xs text-sea-ink-soft leading-relaxed">
                {item.description}
              </p>
            )}
          </>
        )

        if (safeHref) {
          return (
            <a
              key={idx}
              href={safeHref}
              target="_blank"
              rel="noopener noreferrer"
              className="block border border-line bg-surface p-4 transition-colors hover:bg-surface-muted cursor-pointer"
            >
              {content}
            </a>
          )
        }

        return (
          <div key={idx} className="block border border-line bg-surface p-4">
            {content}
          </div>
        )
      })}
    </div>
  )
}
