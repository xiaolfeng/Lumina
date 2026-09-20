import { MarkdownLite } from '@lumina/components/markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwCalloutProps } from '../types'

const LEVEL_STYLES: Record<
  NonNullable<LpwCalloutProps['level']>,
  { border: string; bg: string; title: string }
> = {
  info: {
    border: 'border-lagoon',
    bg: 'bg-lagoon/5',
    title: 'text-lagoon-deep',
  },
  success: {
    border: 'border-kicker',
    bg: 'bg-kicker/5',
    title: 'text-kicker',
  },
  warning: {
    border: 'border-palm',
    bg: 'bg-palm/5',
    title: 'text-palm',
  },
  error: {
    border: 'border-destructive',
    bg: 'bg-destructive/5',
    title: 'text-destructive',
  },
}

export const CalloutBlock: React.FC<LpwBlockSlotProps<LpwCalloutProps>> = ({
  props,
}) => {
  const level = props.level ?? 'info'
  const style = LEVEL_STYLES[level]

  return (
    <div
      data-testid="callout-block"
      data-level={level}
      className={`my-4 border-l-4 p-4 text-sm text-sea-ink ${style.border} ${style.bg}`}
    >
      {props.title && (
        <div className={`mb-1 font-semibold ${style.title}`}>{props.title}</div>
      )}
      <div className="leading-relaxed">
        <MarkdownLite>{props.content}</MarkdownLite>
      </div>
    </div>
  )
}
