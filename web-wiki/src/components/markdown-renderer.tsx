/**
 * Wiki Markdown 渲染器
 *
 * 使用 @lumina/components/markdown 共享渲染原语：
 * - remark-gfm：表格、删除线、任务列表
 * - remark-math + rehype-katex：数学公式
 * - rehype-highlight：代码语法高亮（微明 hljs 主题，透明背景）
 * - rehype-mermaid：Mermaid 图表（inline-svg，按需加载）
 *
 * 安全：react-markdown 默认不渲染 raw HTML（不使用 rehype-raw），无需 rehype-sanitize。
 */
import { Markdown, proseArticle } from '@lumina/components/markdown'

export function MarkdownRenderer({ content }: { content: string }) {
  return <Markdown className={proseArticle}>{content}</Markdown>
}
