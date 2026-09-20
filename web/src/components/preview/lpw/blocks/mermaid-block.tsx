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
      className="my-4 flex flex-col items-center"
    >
      <div className="w-full flex justify-center">
        <Markdown>{fenced}</Markdown>
      </div>
      {props.caption && (
        <figcaption className="mt-1.5 text-center text-xs text-sea-ink-soft">
          {props.caption}
        </figcaption>
      )}
    </figure>
  )
}
