/**
 * fenced-components.tsx
 *
 * React components for fenced blocks (:::callout, :::card, :::steps, :::step).
 *
 * These components are mapped via react-markdown's `components` prop
 * to custom hast elements produced by remark-fenced-blocks.
 */

import type { ReactNode } from 'react'
import { cn } from '../lib/utils'

// ── Callout ──────────────────────────────────────────────────────────

interface CalloutProps {
  type?: 'info' | 'warning' | 'tip' | 'danger'
  children?: ReactNode
}

export function Callout({ type = 'info', children }: CalloutProps) {
  const variants = {
    info: 'border-lagoon/20 bg-lagoon/5 text-lagoon-deep',
    warning: 'border-palm/20 bg-palm/5 text-palm',
    tip: 'border-kicker/20 bg-kicker/5 text-kicker',
    danger: 'border-destructive/20 bg-destructive/5 text-destructive',
  }

  return (
    <div className={cn('my-4 rounded-lg border-l-4 p-4', variants[type])}>
      {children}
    </div>
  )
}

// ── Card ─────────────────────────────────────────────────────────────

interface CardProps {
  title?: string
  href?: string
  children?: ReactNode
}

// sanitizeHref 采用协议白名单（S-01 同源修复）：仅放行 http(s)/mailto 与纯相对引用
// （/、./、../、# 开头或不含冒号的相对文件名）。浏览器解析 URL 会剔除 scheme 内的
// TAB/LF/CR 等控制字符，黑名单前缀匹配可被穿透，因此含控制字符整体拒绝。
function sanitizeHref(href?: string): string | undefined {
  if (!href) return undefined
  const value = href.trim()
  // eslint-disable-next-line no-control-regex
  if (/[\u0000-\u001f\u007f]/.test(value)) return undefined
  if (/^(https?:|mailto:)/i.test(value)) return value
  if (/^[/?#]/.test(value)) return value
  if (/^\.\.?\//.test(value)) return value
  if (!value.includes(':')) return value
  return undefined
}

export function Card({ title, href, children }: CardProps) {
  const safeHref = sanitizeHref(href)
  const className = cn(
    'my-4 block rounded-lg border border-line bg-surface p-4 transition-colors',
    safeHref && 'hover:bg-surface-strong',
  )

  if (safeHref) {
    return (
      <a href={safeHref} className={className}>
        {title && (
          <h3 className="mb-2 text-lg font-semibold text-sea-ink">{title}</h3>
        )}
        {children}
      </a>
    )
  }

  return (
    <div className={className}>
      {title && (
        <h3 className="mb-2 text-lg font-semibold text-sea-ink">{title}</h3>
      )}
      {children}
    </div>
  )
}

// ── Steps ────────────────────────────────────────────────────────────

export function Steps({ children }: { children?: ReactNode }) {
  return (
    <ol className="my-4 list-decimal space-y-2 pl-6">
      {children}
    </ol>
  )
}

// ── Step ─────────────────────────────────────────────────────────────

export function Step({ children }: { children?: ReactNode }) {
  return <li className="text-sea-ink">{children}</li>
}
