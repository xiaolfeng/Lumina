import { Markdown } from '../../markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwMermaidProps } from '../types'
import { MermaidViewport } from './mermaid-viewport'

export const MermaidBlock: React.FC<LpwBlockSlotProps<LpwMermaidProps>> = ({
  props,
}) => {
  const fenced = '```mermaid\n' + props.content.trim() + '\n```'

  return (
    <figure data-testid="mermaid-block" className="font-sans">
      <MermaidViewport>
        <Markdown>{fenced}</Markdown>
      </MermaidViewport>
      {props.caption && (
        <figcaption className="mt-3 text-center text-xs font-serif italic text-sea-ink-soft">
          {props.caption}
        </figcaption>
      )}
    </figure>
  )
}
