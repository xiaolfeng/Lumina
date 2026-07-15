/**
 * Wiki 侧边栏导航组件（shadcn Sidebar inset + 整块淡入动画）
 *
 * 功能特性：
 * - 从 manifest API 获取导航结构（TanStack Query 自动管理）
 * - 使用 shadcn/ui Sidebar variant="inset" 布局（自动处理移动端 Sheet 行为）
 * - 整块淡入（sidebarBlockFade）—— 不做逐项交错，避免路由切换重放序列动画
 * - 树形嵌套渲染：子目录通过左侧竖线（guide line）+ 缩进表达层级关系
 * - 当前页面路径高亮显示
 * - 底部 Powered-by Lumina
 */
import { useEffect, useRef, useState } from 'react'
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
import { cn } from '@lumina/components/lib/utils'
import { sidebarBlockFade } from '@lumina/components/motion'
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
  const hasAutoExpanded = useRef(false)

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
  const projectName = manifest?.project_name ?? 'Wiki'

  useEffect(() => {
    if (hasAutoExpanded.current) return
    if (!currentPagePath || navEntries.length === 0) return
    const dirsToExpand = new Set<string>()
    const pathParts = currentPagePath.split('/').filter(Boolean)
    let currentPath = ''
    for (const part of pathParts.slice(0, -1)) {
      currentPath += (currentPath ? '/' : '') + part
      dirsToExpand.add(currentPath)
    }
    if (dirsToExpand.size > 0) {
      setExpandedDirs(dirsToExpand)
    }
    hasAutoExpanded.current = true
  }, [currentPagePath, navEntries.length])

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

  const renderNavItem = (entry: WikiNavItem, depth: number) => {
    const dirKey = entry.path || entry.title
    const isExpanded = expandedDirs.has(dirKey)
    const isDirectory =
      entry.children !== undefined && entry.children.length > 0
    const isActive = !isDirectory && entry.path === currentPagePath

    return (
      <SidebarMenuItem key={dirKey}>
        {isDirectory ? (
          <SidebarMenuButton
            isActive={false}
            tooltip={entry.title}
            className={cn(
              depth > 0 && 'text-[13px]',
              isExpanded
                ? 'bg-accent/50 text-lagoon font-medium'
                : '',
            )}
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
              <FolderOpen
                className={cn(
                  'size-4',
                  depth === 0 ? 'text-lagoon' : 'text-sea-ink-soft',
                )}
              />
            ) : (
              <FolderClosed
                className={cn(
                  'size-4',
                  depth === 0 ? 'text-muted-foreground' : 'text-sea-ink-soft/70',
                )}
              />
            )}
            <span className="truncate">{entry.title}</span>
          </SidebarMenuButton>
        ) : (
          <SidebarMenuButton
            asChild
            isActive={isActive}
            tooltip={entry.title}
            className={cn(
              depth > 0 && 'text-[13px]',
              isActive
                ? 'bg-chip-bg text-lagoon border border-chip-line font-medium'
                : depth === 0
                  ? ''
                  : 'text-sea-ink-soft',
            )}
          >
            <Link
              to="/wiki/$wikiId/$"
              params={{ wikiId, _splat: entry.path }}
            >
              <FileText
                className={cn(
                  'size-4',
                  depth > 0 && !isActive && 'text-sea-ink-soft/60',
                )}
              />
              <span className="truncate">{entry.title}</span>
            </Link>
          </SidebarMenuButton>
        )}

        {/* 子项：嵌套在 li 内的 ul，带左侧竖线 guide line 表达树形层级 */}
        {isExpanded && entry.children && entry.children.length > 0 && (
          <SidebarMenu className="mt-0.5 min-w-0 gap-0 overflow-hidden border-l border-line/60 pl-1.5 [&>li]:min-w-0">
            {entry.children.map((child) => renderNavItem(child, depth + 1))}
          </SidebarMenu>
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
        variants={sidebarBlockFade}
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
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
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu className="min-w-0">
                {isLoading && (
                  <div className="flex items-center justify-center gap-2 py-6 text-sm text-sea-ink-soft">
                    <Loader2 className="size-4 animate-spin text-lagoon" />
                    <span>加载中...</span>
                  </div>
                )}

                {error && (
                  <div className="mx-1 rounded-md bg-destructive/10 p-3 text-xs text-destructive">
                    {error instanceof Error ? error.message : '加载导航失败'}
                  </div>
                )}

                {!isLoading &&
                  !error &&
                  navEntries.length > 0 &&
                  navEntries.map((entry) => renderNavItem(entry, 0))}

                {!isLoading && !error && navEntries.length === 0 && (
                  <div className="py-8 text-center text-sm text-sea-ink-soft">
                    暂无页面
                  </div>
                )}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter className="border-t border-line">
          <SidebarMenu>
            <SidebarMenuItem>
              <div className="flex items-center gap-2 px-2 py-2 text-xs text-sea-ink-soft">
                <Sparkles className="size-3.5 shrink-0 text-lagoon" />
                <span>
                  由 <span className="font-medium text-sea-ink">Lumina · 微明</span> 驱动
                </span>
              </div>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </motion.div>
    </Sidebar>
  )
}
