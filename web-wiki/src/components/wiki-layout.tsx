/**
 * Wiki 布局组件
 *
 * 功能特性：
 * - 左侧固定宽度侧边栏（WikiSidebar）
 * - 右侧自适应内容区域
 * - 响应式设计：移动端侧边栏可通过按钮切换显示/隐藏
 * - 侧边栏状态管理（展开/折叠）
 */
import { useState } from 'react'
import { Menu } from 'lucide-react'
import { WikiSidebar } from '#/components/wiki-sidebar'
import { MarkdownRenderer } from '#/components/markdown-renderer'

interface WikiLayoutProps {
  wikiId: string
  currentPagePath?: string
  content: string
  title?: string
  children?: React.ReactNode
}

export function WikiLayout({
  wikiId,
  currentPagePath = '',
  content,
  title,
  children,
}: WikiLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="wiki-layout flex h-screen overflow-hidden bg-bg-base">
      {/* 侧边栏 */}
      <WikiSidebar
        wikiId={wikiId}
        currentPagePath={currentPagePath}
        isOpen={sidebarOpen}
        onToggle={() => setSidebarOpen(!sidebarOpen)}
      />

      {/* 主内容区 */}
      <main className="main-content flex flex-1 flex-col overflow-hidden">
        {/* 顶部工具栏 */}
        <header className="flex items-center justify-between border-b border-line bg-header-bg px-4 py-3 backdrop-blur-sm">
          <div className="flex items-center gap-3">
            {/* 移动端菜单按钮 */}
            <button
              onClick={() => setSidebarOpen(true)}
              className="inline-flex items-center justify-center rounded-md p-2 hover:bg-accent md:hidden"
              aria-label="打开导航"
            >
              <Menu className="h-5 w-5" />
            </button>

            {/* 页面标题 */}
            {title && (
              <h1 className="display-title text-lg font-semibold text-sea-ink">
                {title}
              </h1>
            )}
          </div>
        </header>

        {/* 内容滚动区 */}
        <div className="content-scroll flex-1 overflow-y-auto">
          <article className="mx-auto max-w-4xl px-6 py-8 lg:px-8">
            {children ?? <MarkdownRenderer content={content} />}
          </article>
        </div>
      </main>
    </div>
  )
}
