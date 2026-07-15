'use client'

/**
 * TableOfContents — Markdown 页面内目录（右侧 sticky 导航）
 *
 * 功能：
 * - 从 Markdown 原文提取 h2/h3 标题（h1 视为页面标题，不重复收录）
 * - scrollspy：scroll 事件 + getBoundingClientRect，高亮当前阅读章节
 * - 阅读进度百分比（基于 scrollY / scrollHeight）
 * - 点击锚点：平滑滚动到目标标题（依赖 rehype-slug 生成的 id）
 * - 标题为空（无 h2/h3）时整块不渲染
 *
 * 用法：
 *   <TableOfContents content={markdownText} />
 *
 * 依赖：Markdown 渲染器必须启用 rehype-slug（已在 @lumina/components/markdown 中集成）。
 */
import { useEffect, useMemo, useState } from 'react'
import { cn } from '../lib/utils'

export interface TocItem {
  id: string
  text: string
  level: 2 | 3
}

export interface TableOfContentsProps {
  content: string
  className?: string
}

export function extractTocItems(content: string): TocItem[] {
  const items: TocItem[] = []
  const stripped = content
    .replace(/```[\s\S]*?```/g, '')
    .replace(/~~~[\s\S]*?~~~/g, '')

  for (const line of stripped.split('\n')) {
    const m = /^(#{2,3})\s+(.+?)\s*#*$/.exec(line)
    if (!m) continue
    const level = m[1].length as 2 | 3
    const raw = m[2]
      .replace(/`([^`]+)`/g, '$1')
      .replace(/\*\*([^*]+)\*\*/g, '$1')
      .replace(/\*([^*]+)\*/g, '$1')
      .trim()
    if (!raw) continue
    items.push({ id: slugify(raw), text: raw, level })
  }
  return items
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

export function TableOfContents({ content, className }: TableOfContentsProps) {
  const items = useMemo(() => extractTocItems(content), [content])
  const [activeId, setActiveId] = useState('')
  const [progress, setProgress] = useState(0)

  useEffect(() => {
    if (items.length === 0) return

    let headings: HTMLElement[] = []
    let ticking = false
    const THRESHOLD = 100

    const queryHeadings = () => {
      headings = items
        .map((it) => document.getElementById(it.id))
        .filter((el): el is HTMLElement => el !== null)
    }

    const update = () => {
      ticking = false
      if (headings.length === 0) return
      let current = headings[0].id
      for (const h of headings) {
        if (h.getBoundingClientRect().top <= THRESHOLD) {
          current = h.id
        }
      }
      setActiveId(current)

      const scrollHeight =
        document.documentElement.scrollHeight - window.innerHeight
      const pct = scrollHeight > 0 ? (window.scrollY / scrollHeight) * 100 : 0
      setProgress(Math.min(100, Math.max(0, Math.round(pct))))
    }

    const onScroll = () => {
      if (!ticking) {
        requestAnimationFrame(update)
        ticking = true
      }
    }

    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', onScroll)

    // MutationObserver：rehype-slug 注入 id / 图片懒加载撑高 / SPA 页面切换
    // 时，重新探测 headings 并刷新高亮与进度。覆盖 rehype-slug 与 TOC effect
    // 的竞态（即便 slug 已静态加载，图片/mermaid 异步渲染仍会改变布局）。
    const observer = new MutationObserver(() => {
      const prevCount = headings.length
      queryHeadings()
      if (headings.length !== prevCount || headings.length > 0) {
        update()
      }
    })
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['id', 'class'],
    })

    // 首次探测 + 刷新
    queryHeadings()
    const timer = requestAnimationFrame(update)

    return () => {
      window.removeEventListener('scroll', onScroll)
      window.removeEventListener('resize', onScroll)
      observer.disconnect()
      cancelAnimationFrame(timer)
    }
  }, [items])

  const handleClick = (
    e: React.MouseEvent<HTMLAnchorElement>,
    id: string,
  ) => {
    e.preventDefault()
    const el = document.getElementById(id)
    if (!el) return
    const top = el.getBoundingClientRect().top + window.scrollY - 80
    window.scrollTo({ top, behavior: 'smooth' })
    if (typeof history !== 'undefined') {
      history.replaceState(null, '', `#${id}`)
    }
    setActiveId(id)
  }

  if (items.length === 0) return null

  return (
    <nav
      aria-label="页面目录"
      className={cn('flex flex-col gap-3 text-sm', className)}
    >
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-sea-ink-soft">
            阅读进度
          </span>
          <span className="text-xs font-semibold tabular-nums text-lagoon">
            {progress}%
          </span>
        </div>
        <div className="h-1 w-full overflow-hidden rounded-full bg-line">
          <div
            className="h-full rounded-full bg-lagoon transition-[width] duration-300 ease-out"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <p className="text-xs font-semibold uppercase tracking-wider text-sea-ink-soft">
          本页目录
        </p>
        <ul className="flex flex-col gap-0.5 border-l border-line">
          {items.map((it) => {
            const isActive = activeId === it.id
            return (
              <li key={it.id}>
                <a
                  href={`#${it.id}`}
                  onClick={(e) => handleClick(e, it.id)}
                  className={cn(
                    'block cursor-pointer border-l-2 py-1 leading-snug transition-colors',
                    it.level === 2 ? 'pl-3' : 'pl-6 text-xs',
                    isActive
                      ? 'border-lagoon font-medium text-lagoon'
                      : 'border-transparent text-sea-ink-soft hover:border-line hover:text-lagoon-deep',
                  )}
                >
                  {it.text}
                </a>
              </li>
            )
          })}
        </ul>
      </div>
    </nav>
  )
}
