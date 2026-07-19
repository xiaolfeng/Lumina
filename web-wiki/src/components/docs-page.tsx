/**
 * DocsPage 组件 — 三栏 Wiki 文档页面布局
 *
 * 替换旧的 WikiLayout，修复 motion key 导致的全量重载问题。
 *
 * 三栏布局（xl 屏以上）：
 *   [SidebarProvider + PageTreeSidebar | Article (flex-1) | TOC (sticky)]
 * 小屏（< xl）自动隐藏 TOC，仅保留 Sidebar + Article。
 *
 * 关键修复：motion.div 的 key 从 currentPagePath 改为 wikiId，
 * 使得同一 Wiki 内页面切换时 Sidebar 不会重载（保持展开状态），
 * 只有切换不同 Wiki 时才触发 remount。
 */
import type { ReactNode } from 'react'
import { AnimatePresence, motion } from 'motion/react'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '@lumina/components/ui/sidebar'
import { mainSlideIn } from '@lumina/components/motion'
import { PageTreeSidebar } from '#/components/page-tree-sidebar'
import { Breadcrumb } from '#/components/breadcrumb'
import { PrevNext } from '#/components/prev-next'
import { Toc } from '#/components/toc'
import { WikiSearch } from '#/components/search'
import type { PageResponse } from '#/lib/api-client'
import type { PageTree } from '#/lib/source'

interface DocsPageProps {
  wikiId: string
  tree?: PageTree
  pageData?: PageResponse
  children: ReactNode
}

function formatLastUpdated(timestamp: number): string {
  const date = new Date(timestamp * 1000)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export function DocsPage({ wikiId, tree, pageData, children }: DocsPageProps) {
  const currentPagePath = pageData?.path ?? ''

  return (
    <SidebarProvider>
      {/* Sidebar 在 motion.div 之外，避免页面切换时重载 */}
      <PageTreeSidebar wikiId={wikiId} currentPagePath={currentPagePath} />

      <SidebarInset>
        {/* 顶部 header：SidebarTrigger + 页面标题 + WikiSearch */}
        <header className="flex h-16 shrink-0 items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          {pageData?.title && (
            <h1 className="text-lg font-semibold text-sea-ink">
              {pageData.title}
            </h1>
          )}
          <div className="ml-auto">
            <WikiSearch wikiId={wikiId} leaves={tree?.leaves ?? []} />
          </div>
        </header>

        {/* 三栏 flex：正文 + TOC；TOC 在 motion.div 之外保证 sticky 正常 */}
        <div className="mx-auto flex w-full max-w-7xl gap-8 px-6 py-8 lg:px-8">
          {/* Article 区域：motion 仅包裹内容，key=wikiId 避免同 wiki 切换时重载 */}
          <AnimatePresence mode="wait">
            <motion.div
              key={wikiId}
              variants={mainSlideIn}
              initial="hidden"
              animate="visible"
              exit="exit"
              className="min-w-0 flex-1"
            >
              <article>
                {/* Header：标题 + 描述 + 面包屑 + 更新时间 */}
                <header className="mb-8">
                  {pageData?.title && (
                    <h1 className="mb-2 text-3xl font-bold tracking-tight text-sea-ink">
                      {pageData.title}
                    </h1>
                  )}
                  {pageData?.description && (
                    <p className="mb-4 text-lg text-sea-ink-soft">
                      {pageData.description}
                    </p>
                  )}
                  <div className="flex flex-wrap items-center gap-4">
                    <Breadcrumb
                      items={pageData?.breadcrumb}
                      wikiId={wikiId}
                    />
                    {pageData?.last_updated && (
                      <span className="text-sm text-sea-ink-soft/60">
                        更新于 {formatLastUpdated(pageData.last_updated)}
                      </span>
                    )}
                  </div>
                </header>

                {/* Body */}
                <div className="prose prose-slate max-w-none">
                  {children}
                </div>

                {/* Footer：上一页/下一页 */}
                <footer className="mt-12">
                  <PrevNext
                    prev={pageData?.prev ?? null}
                    next={pageData?.next ?? null}
                    wikiId={wikiId}
                  />
                </footer>
              </article>
            </motion.div>
          </AnimatePresence>

          {/* TOC 右栏：xl 屏以上显示 */}
          <aside className="hidden w-56 shrink-0 xl:block">
            <div className="sticky top-8">
              <Toc content={pageData?.content || ''} />
            </div>
          </aside>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
