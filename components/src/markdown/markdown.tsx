import type { ComponentPropsWithoutRef } from 'react'
import { lazy, Suspense } from 'react'
import ReactMarkdown from 'react-markdown'
import { MarkdownLite } from './markdown-lite'

const MarkdownMermaid = lazy(() => import('./markdown-mermaid'))

/**
 * 统一 Markdown 渲染组件。
 *
 * 封装 react-markdown + 全套插件：
 * - remark-gfm：表格、删除线、任务列表、自动链接
 * - remark-math + rehype-katex：数学公式（$...$ 行内、$...$ 块级）
 * - rehype-highlight：代码语法高亮（配 styles.css 的微明 hljs 主题）
 * - rehype-mermaid：Mermaid 图表（默认 inline-svg 浏览器端渲染）
 *
 * 优化：当内容不包含 ```mermaid 代码块时，直接复用 MarkdownLite，
 * 避免拖入 mermaid/cytoscape 等超大库；仅在需要时才通过 React.lazy 加载。
 *
 * 用法：<Markdown className={proseQuestion}>{content}</Markdown>
 */
interface MarkdownProps extends Omit<ComponentPropsWithoutRef<typeof ReactMarkdown>, 'children'> {
  children: string
  className?: string
}

export function Markdown({ children, className, ...rest }: MarkdownProps) {
  const hasMermaid = children.includes('```mermaid')

  if (hasMermaid) {
    return (
      <div className={className}>
        <Suspense fallback={<MarkdownLite {...rest}>{children}</MarkdownLite>}>
          <MarkdownMermaid {...rest}>{children}</MarkdownMermaid>
        </Suspense>
      </div>
    )
  }

  return (
    <div className={className}>
      <MarkdownLite {...rest}>{children}</MarkdownLite>
    </div>
  )
}
