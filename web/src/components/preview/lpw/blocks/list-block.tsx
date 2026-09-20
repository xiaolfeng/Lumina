import { MarkdownLite } from '@lumina/components/markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwListProps } from '../types'

export const ListBlock: React.FC<LpwBlockSlotProps<LpwListProps>> = ({
  props,
}) => {
  const style = props.style ?? 'unordered'

  if (style === 'ordered') {
    return (
      <ol className="my-4 list-decimal space-y-1 pl-6 text-sm text-sea-ink">
        {props.items.map((item, index) => (
          <li key={index} className="leading-relaxed">
            <MarkdownLite>{item.content}</MarkdownLite>
          </li>
        ))}
      </ol>
    )
  }

  if (style === 'check') {
    return (
      <ul className="my-4 space-y-1.5 text-sm text-sea-ink">
        {props.items.map((item, index) => {
          const isChecked = Boolean(item.checked)
          return (
            <li key={index} className="flex items-start gap-2 leading-relaxed">
              <span
                aria-checked={isChecked}
                role="checkbox"
                className={`select-none font-mono text-xs ${
                  isChecked ? 'font-bold text-kicker' : 'text-sea-ink-soft/70'
                }`}
              >
                {isChecked ? '✓' : '□'}
              </span>
              <div className="flex-1">
                <MarkdownLite>{item.content}</MarkdownLite>
              </div>
            </li>
          )
        })}
      </ul>
    )
  }

  return (
    <ul className="my-4 list-disc space-y-1 pl-6 text-sm text-sea-ink">
      {props.items.map((item, index) => (
        <li key={index} className="leading-relaxed">
          <MarkdownLite>{item.content}</MarkdownLite>
        </li>
      ))}
    </ul>
  )
}
