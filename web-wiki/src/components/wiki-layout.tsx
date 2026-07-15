/**
 * Wiki 布局组件 — shadcn Sidebar inset 模式
 *
 * 使用 SidebarProvider + SidebarInset + SidebarTrigger 构建布局，
 * 侧边栏展开/折叠状态由 SidebarProvider 内部管理（cookie 驱动）。
 */
import type { ReactNode } from 'react'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '@lumina/components/ui/sidebar'
import { WikiSidebar } from '#/components/wiki-sidebar'
import { MarkdownRenderer } from '#/components/markdown-renderer'

interface WikiLayoutProps {
  wikiId: string
  currentPagePath?: string
  content: string
  title?: string
  children?: ReactNode
}

export function WikiLayout({
  wikiId,
  currentPagePath = '',
  content,
  title,
  children,
}: WikiLayoutProps) {
  return (
    <SidebarProvider>
      <WikiSidebar wikiId={wikiId} currentPagePath={currentPagePath} />

      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          {title && (
            <h1 className="display-title text-lg font-semibold text-sea-ink">
              {title}
            </h1>
          )}
        </header>

        <main className="flex flex-1 flex-col overflow-hidden">
          <div className="flex-1 overflow-y-auto">
            <article className="mx-auto max-w-4xl px-6 py-8 lg:px-8">
              {children ?? <MarkdownRenderer content={content} />}
            </article>
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
