import type { ComponentPropsWithoutRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeHighlight from 'rehype-highlight'
import rehypeKatex from 'rehype-katex'
import rehypeMermaid from 'rehype-mermaid'

/**
 * 统一 Markdown 渲染组件。
 *
 * 封装 react-markdown + 全套插件：
 * - remark-gfm：表格、删除线、任务列表、自动链接
 * - remark-math + rehype-katex：数学公式（$...$ 行内、$$...$$ 块级）
 * - rehype-highlight：代码语法高亮（配 styles.css 的微明 hljs 主题）
 * - rehype-mermaid：Mermaid 图表（默认 inline-svg 浏览器端渲染）
 *
 * 用法：<Markdown className={proseQuestion}>{content}</Markdown>
 */
const remarkPlugins = [remarkGfm, remarkMath]
const rehypePlugins = [
	rehypeHighlight,
	rehypeKatex,
	[rehypeMermaid, { strategy: 'inline-svg' as const }],
]

interface MarkdownProps extends Omit<ComponentPropsWithoutRef<typeof ReactMarkdown>, 'children'> {
	children: string
}

export function Markdown({ children, ...rest }: MarkdownProps) {
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
