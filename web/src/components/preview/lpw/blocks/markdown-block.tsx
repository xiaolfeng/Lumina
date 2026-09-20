import { Markdown, proseArticle } from '@lumina/components/markdown'
import type React from 'react'
import type { LpwBlockSlotProps, LpwMarkdownProps } from '../types'

export const MarkdownBlock: React.FC<LpwBlockSlotProps<LpwMarkdownProps>> = ({
  props,
}) => {
  return (
    <div className={proseArticle}>
      <Markdown>{props.content}</Markdown>
    </div>
  )
}
