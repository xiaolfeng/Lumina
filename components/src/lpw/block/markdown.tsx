import { Markdown, proseArticle } from '../../markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwMarkdownProps } from '../types'

export const MarkdownBlock: React.FC<LpwBlockSlotProps<LpwMarkdownProps>> = ({
  props,
}) => {
  return (
    <div
      className={`${proseArticle} my-5 font-sans leading-relaxed text-sea-ink`}
    >
      <Markdown>{props.content}</Markdown>
    </div>
  )
}
