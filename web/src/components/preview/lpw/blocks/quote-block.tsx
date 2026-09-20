import { MarkdownLite } from '@lumina/components/markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwQuoteProps } from '../types'

export const QuoteBlock: React.FC<LpwBlockSlotProps<LpwQuoteProps>> = ({
  props,
}) => {
  const hasFooter = Boolean(props.author || props.source)

  return (
    <blockquote className="my-4 border-l-2 border-lagoon bg-lagoon/5 p-4 text-sm text-sea-ink">
      <div className="leading-relaxed italic">
        <MarkdownLite>{props.content}</MarkdownLite>
      </div>
      {hasFooter && (
        <footer className="mt-2 text-xs not-italic text-sea-ink-soft">
          — {props.author ?? ''}
          {props.author && props.source
            ? ` · ${props.source}`
            : (props.source ?? '')}
        </footer>
      )}
    </blockquote>
  )
}
