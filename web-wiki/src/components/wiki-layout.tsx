/**
 * Wiki 布局组件 — shadcn Sidebar inset 模式 + TOC 右栏
 *
 * 使用 SidebarProvider + SidebarInset + SidebarTrigger 构建布局，
 * 侧边栏展开/折叠状态由 SidebarProvider 内部管理（cookie 驱动）。
 *
 * 三栏布局（xl 屏以上）：
 *   [Sidebar | Article (flex-1) | TOC (sticky, w-56)]
 * 小屏（< xl）自动隐藏 TOC，仅保留 Sidebar + Article。
 *
 * 关键：TOC 必须在 motion.div 之外 —— motion 的 transform 会创建
 * containing block，破坏 sticky 定位。
 */
import type { ReactNode } from 'react'
import { AnimatePresence, motion } from 'motion/react'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '@lumina/components/ui/sidebar'
import {
  Markdown,
  TableOfContents,
  proseArticle,
} from '@lumina/components/markdown'
import { mainSlideIn } from '@lumina/components/motion'
import { WikiSidebar } from '#/components/wiki-sidebar'

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

        {/* 三栏 flex：正文 + TOC；TOC 在 motion.div 之外保证 sticky 正常 */}
        <div className="mx-auto flex w-full max-w-7xl gap-8 px-6 py-8 lg:px-8">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentPagePath}
              variants={mainSlideIn}
              initial="hidden"
              animate="visible"
              exit="exit"
              className="min-w-0 flex-1"
            >
              <article className={proseArticle}>
                {children ?? <Markdown>{content}</Markdown>}
              </article>
            </motion.div>
          </AnimatePresence>

          {/* TOC 右栏：xl 屏以上显示，无 h2/h3 时 TableOfContents 自动返回 null */}
          <aside className="hidden w-56 shrink-0 xl:block">
            <div className="sticky top-8">
              <TableOfContents content={content} />
            </div>
          </aside>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
