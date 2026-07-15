/**
 * Wiki 侧边栏导航组件（shadcn Sidebar inset + motion 动画）
 *
 * 功能特性：
 * - 从 manifest API 获取导航结构（TanStack Query 自动管理）
 * - 使用 shadcn/ui Sidebar variant="inset" 布局（自动处理移动端 Sheet 行为）
 * - motion 交错入场动画（sidebarStaggerContainer / sidebarItem）
 * - 支持目录展开/折叠交互（递归渲染 WikiNavItem 树）
 * - 当前页面路径高亮显示
 * - 底部 Powered-by Lumina 卡片
 */
import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'motion/react'
import {
  ChevronRight,
  ChevronDown,
  FolderOpen,
  FolderClosed,
  FileText,
  BookOpen,
  Loader2,
  Sparkles,
} from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@lumina/components/ui/sidebar'
import { Card, CardContent } from '@lumina/components/ui/card'
import { sidebarItem, sidebarStaggerContainer } from '@lumina/components/motion'
import { wikiReaderApi } from '#/lib/api-client'
import type { ManifestResponse, WikiNavItem } from '#/lib/api-client'

interface WikiSidebarProps {
  wikiId: string
  currentPagePath?: string
}

export function WikiSidebar({
  wikiId,
  currentPagePath = '',
}: WikiSidebarProps) {
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set())

  const {
    data: manifest,
    isLoading,
    error,
  } = useQuery<ManifestResponse>({
    queryKey: ['wiki-manifest', wikiId],
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    enabled: !!wikiId,
    staleTime: 5 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })

  const navEntries: WikiNavItem[] = manifest?.navigation ?? []
  const homePath = manifest?.home ?? ''
  const projectName = manifest?.project_name ?? 'Wiki'

  // 默认展开包含当前页面的父目录
  if (currentPagePath && expandedDirs.size === 0 && navEntries.length > 0) {
    const dirsToExpand = new Set<string>()
    const pathParts = currentPagePath.split('/').filter(Boolean)
    let currentPath = ''
    for (const part of pathParts.slice(0, -1)) {
      currentPath += (currentPath ? '/' : '') + part
      dirsToExpand.add(currentPath)
    }
    setExpandedDirs(dirsToExpand)
  }

  // 切换目录展开状态
  const toggleDir = (dirPath: string) => {
    setExpandedDirs((prev) => {
      const next = new Set(prev)
      if (next.has(dirPath)) {
        next.delete(dirPath)
      } else {
        next.add(dirPath)
      }
      return next
    })
  }

  // 渲染单个导航项（递归）
  const renderNavItem = (entry: WikiNavItem, depth: number = 0) => {
    const dirKey = entry.path || entry.title
    const isExpanded = expandedDirs.has(dirKey)
    const isDirectory =
      entry.children !== undefined && entry.children.length > 0
    const isActive = !isDirectory && entry.path === currentPagePath

    return (
      <SidebarMenuItem key={dirKey}>
        <motion.div variants={sidebarItem}>
          {isDirectory ? (
            /* ── 目录项：展开/折叠按钮 + 目录名链接 ── */
            <SidebarMenuButton
              isActive={false}
              tooltip={entry.title}
              className={
                isExpanded ? 'bg-accent/50 text-lagoon font-medium' : ''
              }
              onClick={(e) => {
                e.preventDefault()
                toggleDir(dirKey)
              }}
            >
              <span
                className="inline-flex size-4 items-center justify-center rounded hover:bg-muted"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  toggleDir(dirKey)
                }}
              >
                {isExpanded ? (
                  <ChevronDown className="size-3" />
                ) : (
                  <ChevronRight className="size-3" />
                )}
              </span>
              {isExpanded ? (
                <FolderOpen className="size-4 text-lagoon" />
              ) : (
                <FolderClosed className="size-4 text-muted-foreground" />
              )}
              <span className="truncate">{entry.title}</span>
            </SidebarMenuButton>
          ) : (
            /* ── 文件项：页面链接 ── */
            <SidebarMenuButton
              asChild
              isActive={isActive}
              tooltip={entry.title}
              className={
                isActive
                  ? 'bg-chip-bg text-lagoon border border-chip-line font-medium'
                  : ''
              }
            >
              <Link
                to="/wiki/$wikiId/$"
                params={{ wikiId, _splat: entry.path }}
              >
                <FileText className="size-4" />
                <span className="truncate">{entry.title}</span>
              </Link>
            </SidebarMenuButton>
          )}
        </motion.div>

        {/* 子目录递归渲染 */}
        {isExpanded && entry.children && entry.children.length > 0 && (
          <motion.div variants={sidebarItem}>
            {entry.children.map((child) => renderNavItem(child, depth + 1))}
          </motion.div>
        )}
      </SidebarMenuItem>
    )
  }

  return (
    <Sidebar variant="inset">
      <motion.div
        className="flex h-full flex-col"
        initial="hidden"
        animate="visible"
        variants={sidebarStaggerContainer}
      >
        {/* ── 头部：项目名称 ── */}
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <motion.div variants={sidebarItem}>
                <SidebarMenuButton size="lg" className="hover:bg-link-bg-hover">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-lagoon text-foam shadow-sm shadow-hero-a">
                    <BookOpen className="size-4" />
                  </div>
                  <div className="flex flex-col gap-0.5 leading-none">
                    <span className="font-semibold text-sea-ink">
                      {projectName}
                    </span>
                    <span className="text-xs text-sea-ink-soft">Wiki 导航</span>
                  </div>
                </SidebarMenuButton>
              </motion.div>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        {/* ── 导航内容区 ── */}
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {/* 首页链接 */}
                <motion.div variants={sidebarItem}>
                  <SidebarMenuItem>
                    <SidebarMenuButton
                      asChild
                      isActive={!currentPagePath}
                      tooltip="首页"
                      className={
                        !currentPagePath
                          ? 'bg-chip-bg text-lagoon border border-chip-line font-medium'
                          : 'hover:bg-link-bg-hover'
                      }
                    >
                      <Link
                        to={homePath ? '/wiki/$wikiId/$' : '/wiki/$wikiId'}
                        params={
                          homePath ? { wikiId, _splat: homePath } : { wikiId }
                        }
                      >
                        <BookOpen className="size-4" />
                        <span>首页</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </motion.div>

                {/* 加载状态 */}
                {isLoading && (
                  <motion.div variants={sidebarItem}>
                    <div className="flex items-center justify-center gap-2 py-6 text-sm text-sea-ink-soft">
                      <Loader2 className="size-4 animate-spin text-lagoon" />
                      <span>加载中...</span>
                    </div>
                  </motion.div>
                )}

                {/* 错误提示 */}
                {error && (
                  <motion.div variants={sidebarItem}>
                    <div className="mx-1 rounded-md bg-destructive/10 p-3 text-xs text-destructive">
                      {error instanceof Error ? error.message : '加载导航失败'}
                    </div>
                  </motion.div>
                )}

                {/* 导航树 */}
                {!isLoading &&
                  !error &&
                  navEntries.length > 0 &&
                  navEntries.map((entry) => renderNavItem(entry))}

                {/* 空状态 */}
                {!isLoading && !error && navEntries.length === 0 && (
                  <motion.div variants={sidebarItem}>
                    <div className="py-8 text-center text-sm text-sea-ink-soft">
                      暂无页面
                    </div>
                  </motion.div>
                )}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        {/* ── 底部：Powered-by 卡片 ── */}
        <SidebarFooter className="border-t border-line">
          <SidebarMenu>
            <SidebarMenuItem>
              <motion.div variants={sidebarItem}>
                <Card className="border-border/50 bg-surface shadow-none">
                  <CardContent className="flex items-center gap-2 p-3">
                    <Sparkles className="size-4 text-lagoon" />
                    <div className="flex flex-col">
                      <span className="text-xs font-medium text-sea-ink">
                        由 Lumina · 微明 驱动
                      </span>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </motion.div>
    </Sidebar>
  )
}
