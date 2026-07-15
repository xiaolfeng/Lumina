import type { ComponentPropsWithoutRef } from 'react'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'

/**
 * 轻量 Markdown 渲染组件（不含 mermaid / highlight / katex）。
 *
 * 封装 react-markdown + 常用插件：
 * - remark-gfm：表格、删除线、任务列表、自动链接
 * - remark-math：数学公式语法解析（仅 AST 层）
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
  const [rehypePlugins, setRehypePlugins] = useState<any[]>([])

  useEffect(() => {
    let cancelled = false
    Promise.all([
      import('rehype-highlight'),
      import('rehype-katex'),
    ]).then(([hl, katex]) => {
      if (!cancelled) {
        setRehypePlugins([hl.default, katex.default])
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
