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

export function Card({ title, href, children }: CardProps) {
  const className = cn(
    'my-4 block rounded-lg border border-line bg-surface p-4 transition-colors',
    href && 'hover:bg-surface-strong',
  )

  if (href) {
    return (
      <a href={href} className={className}>
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
