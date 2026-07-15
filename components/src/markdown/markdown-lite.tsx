import type { ComponentPropsWithoutRef } from 'react'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeSlug from 'rehype-slug'

/**
 * 轻量 Markdown 渲染组件（不含 mermaid）。
 *
 * 封装 react-markdown + 常用插件：
 * - remark-gfm：表格、删除线、任务列表、自动链接
 * - remark-math：数学公式语法解析（仅 AST 层）
 * - rehype-slug：标题 id 生成（静态，~3KB，保证 TOC scrollspy 能及时拿到锚点）
 *
 * rehype-highlight（代码高亮，拖入 highlight.js ~2.7MB）和
 * rehype-katex（公式渲染，拖入 katex ~280KB）通过动态 import
 * 按需加载，避免首屏 vendor chunk 过大。
 *
 * 用法：<MarkdownLite className={proseQuestion}>{content}</MarkdownLite>
 */
interface MarkdownLiteProps extends Omit<ComponentPropsWithoutRef<typeof ReactMarkdown>, 'children'> {
  children: string
  className?: string
}

export function MarkdownLite({ children, className, ...rest }: MarkdownLiteProps) {
  // rehype-slug 静态加载：必须同步可用，否则标题无 id，TOC scrollspy 失效
  const [rehypePlugins, setRehypePlugins] = useState<any[]>([rehypeSlug])

  useEffect(() => {
    let cancelled = false
    Promise.all([
      import('rehype-highlight'),
      import('rehype-katex'),
    ]).then(([hl, katex]) => {
      if (!cancelled) {
        setRehypePlugins([rehypeSlug, hl.default, katex.default])
      }
    })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className={className}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={rehypePlugins as never}
        {...rest}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
