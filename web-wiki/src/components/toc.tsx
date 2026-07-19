import { useEffect, useMemo, useState } from 'react'
import { extractToc } from '#/lib/frontmatter'

function getIndentClass(depth: number): string {
  if (depth <= 2) return 'ps-3'
  if (depth === 3) return 'ps-6'
  return 'ps-8'
}

export function Toc({ content }: { content: string }) {
  const items = useMemo(() => extractToc(content), [content])
  const [activeSlug, setActiveSlug] = useState<string>('')

  useEffect(() => {
    if (items.length === 0) return

    const headings = items
      .map((item) => document.getElementById(item.slug))
      .filter((el): el is HTMLElement => el !== null)

    if (headings.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .map((e) => e.target.id)

        if (visible.length > 0) {
          setActiveSlug(visible[0])
        }
      },
      {
        rootMargin: '-80px 0px -70% 0px',
        threshold: 0,
      },
    )

    for (const h of headings) {
      observer.observe(h)
    }

    return () => {
      observer.disconnect()
    }
  }, [items])

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, slug: string) => {
    e.preventDefault()
    const el = document.getElementById(slug)
    if (!el) return
    const top = el.getBoundingClientRect().top + window.scrollY - 80
    window.scrollTo({ top, behavior: 'smooth' })
    if (typeof history !== 'undefined') {
      history.replaceState(null, '', `#${slug}`)
    }
    setActiveSlug(slug)
  }

  if (items.length === 0) return null

  return (
    <nav aria-label="页面目录" className="flex flex-col gap-3 text-sm">
      <div className="flex flex-col gap-1">
        <p className="text-xs font-semibold uppercase tracking-wider text-sea-ink-soft">
          本页目录
        </p>
        <ul className="flex flex-col gap-0.5 border-l border-line">
          {items.map((item) => {
            const isActive = activeSlug === item.slug
            return (
              <li key={item.slug}>
                <a
                  href={`#${item.slug}`}
                  onClick={(e) => handleClick(e, item.slug)}
                  className={[
                    'block cursor-pointer border-l-2 py-1 leading-snug transition-colors',
                    getIndentClass(item.depth),
                    isActive
                      ? 'border-lagoon font-medium text-lagoon'
                      : 'border-transparent text-sea-ink-soft hover:border-line hover:text-lagoon-deep',
                  ].join(' ')}
                >
                  {item.title}
                </a>
              </li>
            )
          })}
        </ul>
      </div>
    </nav>
  )
}
