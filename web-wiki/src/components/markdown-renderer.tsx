/**
 * Wiki Markdown 渲染器
 *
 * 功能特性：
 * - Markdown → HTML 转换（react-markdown + remark-gfm）
 * - 代码高亮（highlight.js GitHub Dark 主题）
 * - Mermaid 图表渲染（pre-mermaid 策略）
 * - XSS 防护（rehype-sanitize，必须放在最前）
 *
 * 安全约束：
 * - 禁止使用 dangerouslySetInnerHTML
 * - rehype-sanitize 必须在所有 rehype 插件之前
 */
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import rehypeMermaid from 'rehype-mermaid'
import rehypeSanitize from 'rehype-sanitize'
import { useEffect } from 'react'

// 引入 highlight.js GitHub Dark 主题
import 'highlight.js/styles/github-dark.css'

interface MermaidWindow extends Window {
  mermaid?: {
    initialize: (config: Record<string, unknown>) => void
  }
}

export function MarkdownRenderer({ content }: { content: string }) {
  // 初始化 Mermaid 主题配置
  useEffect(() => {
    // 配置 Mermaid 使用与微明色盘兼容的主题
    const win = window as MermaidWindow
    if (typeof window !== 'undefined' && win.mermaid) {
      win.mermaid.initialize({
        startOnLoad: false,
        theme: 'dark',
        themeVariables: {
          primaryColor: '#c9883a',
          primaryTextColor: '#e8ddd0',
          primaryBorderColor: '#c9883a',
          lineColor: '#a89a8a',
          secondaryColor: '#2a2420',
          tertiaryColor: '#1a1512',
          background: '#13100e',
          mainBkg: '#1a1512',
          nodeBorder: '#c9883a',
          clusterBkg: '#2a2420',
          titleColor: '#e8ddd0',
          edgeLabelBackground: '#1a1512',
        },
        securityLevel: 'strict', // 安全模式：禁用 HTML 注入
        fontFamily: 'Manrope, ui-sans-serif, system-ui, sans-serif',
      })
    }
  }, [])

  return (
    <div className="wiki-content prose prose-lg max-w-none dark:prose-invert">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[
          rehypeSanitize, // XSS 防护（必须！必须在最前面）
          rehypeHighlight, // 代码高亮
          [rehypeMermaid, { strategy: 'pre-mermaid' }], // Mermaid 图表
        ]}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
