import type { ComponentPropsWithoutRef } from 'react'
import { MarkdownHooks } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeSlug from 'rehype-slug'
import rehypeHighlight from 'rehype-highlight'
import rehypeKatex from 'rehype-katex'
import rehypeMermaid from 'rehype-mermaid'
import { remarkFencedBlocks } from './remark-fenced-blocks'
import { Callout, Card, Steps, Step } from './fenced-components'

/**
 * 含 Mermaid 图表支持的 Markdown 渲染组件。
 *
 * 在 MarkdownLite 基础上额外注入 rehype-mermaid，
 * 支持 ```mermaid 代码块渲染为 SVG 图表。
 *
 * 由于 rehype-mermaid 会拖入 mermaid/cytoscape（~1000KB），
 * 该组件通过 React.lazy 按需加载，仅在内容包含 mermaid 时才触发 chunk 下载。
 *
 * ⚠️ 必须使用 MarkdownHooks（而非默认的 sync Markdown）：
 * rehype-mermaid 的 inline-svg 渲染是异步的，sync 版内部 processor.runSync
 * 会在遇到异步插件时抛出 `runSync finished async` 错误。
 */
const remarkPlugins = [remarkGfm, remarkMath, remarkFencedBlocks]
const rehypePlugins = [
  rehypeSlug,
  rehypeHighlight,
  rehypeKatex,
  [rehypeMermaid, { strategy: 'inline-svg' as const }],
]

interface MarkdownMermaidProps extends Omit<ComponentPropsWithoutRef<typeof MarkdownHooks>, 'children'> {
  children: string
  className?: string
}

export default function MarkdownMermaid({ children, className, components: userComponents, ...rest }: MarkdownMermaidProps) {
  return (
    <div className={className}>
      <MarkdownHooks
        remarkPlugins={remarkPlugins}
        rehypePlugins={rehypePlugins as never}
        components={{
          callout: Callout,
          card: Card,
          steps: Steps,
          step: Step,
          ...(userComponents || {}),
        } as never}
        {...rest}
      >
        {children}
      </MarkdownHooks>
    </div>
  )
}
