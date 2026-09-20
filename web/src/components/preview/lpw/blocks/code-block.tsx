import { Badge } from '@lumina/components/ui/badge'
import React, { useMemo } from 'react'
import type { LpwBlockSlotProps, LpwCodeProps } from '../types'

export const CodeBlock: React.FC<LpwBlockSlotProps<LpwCodeProps>> = ({
  props,
}) => {
  const lines = useMemo(() => props.content.split('\n'), [props.content])
  const highlightSet = useMemo(
    () => new Set(props.highlightLines ?? []),
    [props.highlightLines],
  )
  const showHeader = Boolean(props.filename || props.language)
  const showLineNumbers = Boolean(props.showLineNumbers)

  return (
    <div
      data-testid="code-block"
      className="my-4 border border-line bg-surface text-xs"
    >
      {showHeader && (
        <div className="flex items-center justify-between border-b border-line bg-surface-muted/30 px-3 py-1.5">
          <span className="font-mono text-sea-ink-soft">
            {props.filename ?? ''}
          </span>
          {props.language && (
            <Badge
              variant="outline"
              className="font-mono text-[10px] uppercase"
            >
              {props.language}
            </Badge>
          )}
        </div>
      )}

      <pre className="overflow-x-auto p-3 font-mono text-xs leading-relaxed text-sea-ink">
        <code>
          {lines.map((line, idx) => {
            const lineNum = idx + 1
            const isHighlighted = highlightSet.has(lineNum)
            return (
              <div
                key={idx}
                data-line-number={lineNum}
                className={`flex items-center px-1 ${
                  isHighlighted ? 'bg-sand' : ''
                }`}
              >
                {showLineNumbers && (
                  <span className="w-8 shrink-0 select-none pr-3 text-right text-sea-ink-soft/40">
                    {lineNum}
                  </span>
                )}
                <span className="flex-1 whitespace-pre">{line || ' '}</span>
              </div>
            )
          })}
        </code>
      </pre>
    </div>
  )
}
