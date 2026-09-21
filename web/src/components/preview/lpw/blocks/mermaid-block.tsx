import { Markdown } from '@lumina/components/markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwMermaidProps } from '../types'

export const MermaidBlock: React.FC<LpwBlockSlotProps<LpwMermaidProps>> = ({
  props,
}) => {
  const fenced = '```mermaid\n' + props.content.trim() + '\n```'

  return (
    <figure
      data-testid="mermaid-block"
      className="my-8 flex flex-col items-center border border-line/60 bg-surface/30 p-6 shadow-2xs font-sans w-full"
    >
      <div className="w-full flex justify-center">
        <Markdown>{fenced}</Markdown>
      </div>
      {props.caption && (
        <figcaption className="mt-3 text-center text-xs font-serif italic text-sea-ink-soft/90 tracking-wide border-t border-line/40 pt-2 w-full">
          {props.caption}
        </figcaption>
      )}
    </figure>
  )
}
