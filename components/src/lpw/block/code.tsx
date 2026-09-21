import { FileCode } from 'lucide-react'
import React, { useMemo } from 'react'
import { Badge } from '../../ui/badge'
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
  const lineNoWidth = lines.length >= 1000 ? 'w-10' : 'w-8'

  return (
    <div
      data-testid="code-block"
      className="my-6 border border-line bg-surface/40 text-xs shadow-2xs font-mono"
    >
      {showHeader && (
        <div className="flex items-center justify-between border-b border-line bg-surface/60 px-4 py-2">
          <div className="flex items-center gap-2 min-w-0">
            <FileCode className="h-3.5 w-3.5 shrink-0 text-sea-ink-soft" />
            <span
              className="font-mono text-xs font-semibold text-sea-ink truncate max-w-[200px] sm:max-w-md"
              title={props.filename}
            >
              {props.filename ?? ''}
            </span>
          </div>
          {props.language && (
            <Badge
              variant="outline"
              className="font-mono text-[10px] uppercase border-line text-sea-ink-soft px-1.5 py-0 shrink-0"
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
                className={`flex items-start min-w-full w-fit px-1.5 py-0.5 ${
                  isHighlighted ? 'bg-sand font-semibold' : ''
                }`}
              >
                {showLineNumbers && (
                  <span
                    className={`${lineNoWidth} shrink-0 select-none pr-3 text-right text-sea-ink-soft/40 font-mono text-[11px]`}
                  >
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
