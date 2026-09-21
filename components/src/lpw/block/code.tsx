import { Badge } from '../../ui/badge'
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
      className="my-6 border border-line bg-surface/40 text-xs shadow-2xs font-mono"
    >
      {showHeader && (
        <div className="flex items-center justify-between border-b border-line bg-surface/60 px-4 py-2">
          <span className="font-mono text-xs font-semibold text-sea-ink">
            {props.filename ?? ''}
          </span>
          {props.language && (
            <Badge
              variant="outline"
              className="font-mono text-[10px] uppercase border-line text-sea-ink-soft px-1.5 py-0"
            >
              {props.language}
            </Badge>
          )}
        </div>
      )}

      <pre className="overflow-x-auto p-4 font-mono text-xs leading-relaxed text-sea-ink bg-surface/20">
        <code>
          {lines.map((line, idx) => {
            const lineNum = idx + 1
            const isHighlighted = highlightSet.has(lineNum)
            return (
              <div
                key={idx}
                data-line-number={lineNum}
                className={`flex items-center px-1.5 py-0.5 ${
                  isHighlighted ? 'bg-sand font-semibold' : ''
                }`}
              >
                {showLineNumbers && (
                  <span className="w-8 shrink-0 select-none pr-3 text-right text-sea-ink-soft/40 font-mono text-[11px]">
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
