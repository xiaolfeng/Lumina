import type { ComponentPropsWithoutRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeHighlight from 'rehype-highlight'
import rehypeKatex from 'rehype-katex'
import rehypeMermaid from 'rehype-mermaid'

/**
 * 含 Mermaid 图表支持的 Markdown 渲染组件。
 *
 * 在 MarkdownLite 基础上额外注入 rehype-mermaid，
 * 支持 ```mermaid 代码块渲染为 SVG 图表。
 *
 * 由于 rehype-mermaid 会拖入 mermaid/cytoscape（~1000KB），
 * 该组件通过 React.lazy 按需加载，仅在内容包含 mermaid 时才触发 chunk 下载。
 */
const remarkPlugins = [remarkGfm, remarkMath]
const rehypePlugins = [
  rehypeHighlight,
  rehypeKatex,
  [rehypeMermaid, { strategy: 'inline-svg' as const }],
]

interface MarkdownMermaidProps extends Omit<ComponentPropsWithoutRef<typeof ReactMarkdown>, 'children'> {
  children: string
}

export default function MarkdownMermaid({ children, ...rest }: MarkdownMermaidProps) {
  return (
    <ReactMarkdown
      remarkPlugins={remarkPlugins}
      rehypePlugins={rehypePlugins as never}
      {...rest}
    >
      {children}
    </ReactMarkdown>
  )
}
