import { MarkdownLite } from '../../markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwCalloutProps } from '../types'

const LEVEL_CONFIGS: Record<
  NonNullable<LpwCalloutProps['level']>,
  { border: string; bg: string; title: string; tag: string }
> = {
  info: {
    border: 'border-lagoon',
    bg: 'bg-lagoon/5',
    title: 'text-lagoon-deep',
    tag: 'CANONICAL DIRECTIVE · 框架规范',
  },
  success: {
    border: 'border-kicker',
    bg: 'bg-kicker/5',
    title: 'text-kicker',
    tag: 'VERIFIED CONFORMANCE · 检验达标',
  },
  warning: {
    border: 'border-palm',
    bg: 'bg-palm/5',
    title: 'text-palm',
    tag: 'SECURITY BOUNDARY · 安全红线',
  },
  error: {
    border: 'border-destructive',
    bg: 'bg-destructive/5',
    title: 'text-destructive',
    tag: 'CRITICAL EXCEPTION · 致命告警',
  },
}

export const CalloutBlock: React.FC<LpwBlockSlotProps<LpwCalloutProps>> = ({
  props,
}) => {
  const level = props.level ?? 'info'
  const config = LEVEL_CONFIGS[level]

  return (
    <div
      data-testid="callout-block"
      data-level={level}
      className={`my-6 p-5 sm:p-6 text-sm text-sea-ink border border-line/50 border-l-4 border-double ${config.border} ${config.bg} shadow-sm relative`}
    >
      {/* 出版物微型分类签 */}
      <div
        className={`font-mono text-[10px] font-bold tracking-widest uppercase mb-1.5 ${config.title}`}
      >
        {config.tag}
      </div>

      {props.title && (
        <div
          className={`mb-2 font-serif text-base font-semibold tracking-tight ${config.title}`}
        >
          {props.title}
        </div>
      )}
      <div className="leading-relaxed font-sans text-sea-ink/90">
        <MarkdownLite>{props.content}</MarkdownLite>
      </div>
    </div>
  )
}
