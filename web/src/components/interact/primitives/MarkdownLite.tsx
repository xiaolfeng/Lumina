import type { ComponentPropsWithoutRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeHighlight from 'rehype-highlight'
import rehypeKatex from 'rehype-katex'

/**
 * 轻量 Markdown 渲染组件（不含 mermaid）。
 *
 * 封装 react-markdown + 常用插件：
 * - remark-gfm：表格、删除线、任务列表、自动链接
 * - remark-math + rehype-katex：数学公式
 * - rehype-highlight：代码语法高亮
 *
 * 不含 rehype-mermaid，因此不会拖入 mermaid/cytoscape 等大库，
 * 适合问题组件等渲染简单文本的场景。
 *
 * 用法：<MarkdownLite className={proseQuestion}>{content}</MarkdownLite>
 */
const remarkPlugins = [remarkGfm, remarkMath]
const rehypePlugins = [rehypeHighlight, rehypeKatex]

interface MarkdownLiteProps extends Omit<ComponentPropsWithoutRef<typeof ReactMarkdown>, 'children'> {
  children: string
}

export function MarkdownLite({ children, ...rest }: MarkdownLiteProps) {
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
