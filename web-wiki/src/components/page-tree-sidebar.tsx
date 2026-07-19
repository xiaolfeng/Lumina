/**
 * Wiki 页面树侧边栏组件（基于 PageTree 递归渲染）
 *
 * 功能特性：
 * - 从 manifest API 获取导航结构，经 buildPageTree 转为 PageTree
 * - 使用 shadcn/ui Sidebar variant="inset"
 * - 整块淡入（sidebarBlockFade）—— 避免逐项交错动画在路由切换时重放
 * - 递归渲染：目录节点（可展开/折叠）+ 叶子节点（Link 导航）+ 分隔符节点
 * - 当前页面路径高亮
 * - 展开状态持久化到 localStorage
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
  SidebarSeparator,
} from '@lumina/components/ui/sidebar'
import { cn } from '@lumina/components/utils'
import { sidebarBlockFade } from '@lumina/components/motion'
import { wikiReaderApi } from '#/lib/api-client'
import { buildPageTree, getIcon } from '#/lib/source'
import type { PageNode, PageTree } from '#/lib/source'

interface PageTreeSidebarProps {
  wikiId: string
  currentPagePath?: string
}

function getExpandedKey(wikiId: string): string {
  return `wiki-sidebar-expanded-${wikiId}`
}

function loadExpanded(wikiId: string): Set<string> {
  try {
    const raw = localStorage.getItem(getExpandedKey(wikiId))
    if (!raw) return new Set()
    const arr = JSON.parse(raw) as string[]
    return new Set(arr)
  } catch {
    return new Set()
  }
}

function saveExpanded(wikiId: string, expanded: Set<string>): void {
  localStorage.setItem(
    getExpandedKey(wikiId),
    JSON.stringify(Array.from(expanded)),
  )
}

function getDirKey(node: PageNode): string {
  return node.path || node.title
}

function shouldExpandByDefault(
  node: PageNode,
  currentPagePath: string,
): boolean {
  if (node.defaultOpen) return true
  if (!currentPagePath || !node.path) return false
  return (
    currentPagePath === node.path ||
    currentPagePath.startsWith(node.path + '/')
  )
}

export function PageTreeSidebar({
  wikiId,
  currentPagePath = '',
}: PageTreeSidebarProps) {
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(() =>
    loadExpanded(wikiId),
  )
  const hasAutoExpanded = useRef(false)

  const {
    data: manifest,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['wiki-manifest', wikiId],
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    enabled: !!wikiId,
    staleTime: 5 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })

  const tree: PageTree | null = manifest ? buildPageTree(manifest) : null
  const projectName = manifest?.meta?.title ?? manifest?.project_name ?? 'Wiki'

  useEffect(() => {
    if (hasAutoExpanded.current) return
    if (!tree) return

    const dirsToExpand = new Set<string>()
    function walk(node: PageNode): void {
      if (node.children && node.children.length > 0) {
        const dirKey = getDirKey(node)
        if (shouldExpandByDefault(node, currentPagePath)) {
          dirsToExpand.add(dirKey)
        }
        for (const child of node.children) {
          walk(child)
        }
      }
    }
    walk(tree.root)

    if (dirsToExpand.size > 0) {
      setExpandedDirs((prev) => new Set([...prev, ...dirsToExpand]))
    }
    hasAutoExpanded.current = true
  }, [tree, currentPagePath])

  useEffect(() => {
    if (!wikiId) return
    saveExpanded(wikiId, expandedDirs)
  }, [wikiId, expandedDirs])

  const toggleDir = (dirKey: string) => {
    setExpandedDirs((prev) => {
      const next = new Set(prev)
      if (next.has(dirKey)) {
        next.delete(dirKey)
      } else {
        next.add(dirKey)
      }
      return next
    })
  }

  const isNodeExpanded = (node: PageNode): boolean => {
    return expandedDirs.has(getDirKey(node))
  }

  const renderNode = (
    node: PageNode,
    depth: number,
    index: number,
  ): React.ReactNode => {
    const dirKey = getDirKey(node)
    const isExpanded = isNodeExpanded(node)
    const isDirectory = node.children && node.children.length > 0
    const isLeaf = !isDirectory && !node.separator
    const isActive = isLeaf && node.path === currentPagePath
    const Icon = getIcon(node.icon)

    if (node.separator) {
      return <SidebarSeparator key={`sep-${index}`} />
    }

    return (
      <SidebarMenuItem key={dirKey}>
        {isDirectory ? (
          <SidebarMenuButton
            isActive={false}
            tooltip={node.title}
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
                  depth === 0
                    ? 'text-lagoon'
                    : 'text-sea-ink-soft',
                )}
              />
            ) : (
              <FolderClosed
                className={cn(
                  'size-4',
                  depth === 0
                    ? 'text-muted-foreground'
                    : 'text-sea-ink-soft/70',
                )}
              />
            )}
            <span className="truncate">{node.title}</span>
          </SidebarMenuButton>
        ) : (
          <SidebarMenuButton
            asChild
            isActive={isActive}
            tooltip={node.title}
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
              params={{ wikiId, _splat: node.path }}
            >
              <Icon
                className={cn(
                  'size-4',
                  depth > 0 && !isActive && 'text-sea-ink-soft/60',
                )}
              />
              <span className="truncate">{node.title}</span>
            </Link>
          </SidebarMenuButton>
        )}

        {isExpanded && node.children && node.children.length > 0 && (
          <SidebarMenu className="mt-0.5 min-w-0 gap-0 overflow-hidden border-l border-line/60 pl-1.5 [&>li]:min-w-0">
            {node.children.map((child, i) =>
              renderNode(child, depth + 1, i),
            )}
          </SidebarMenu>
        )}
      </SidebarMenuItem>
    )
  }

  return (
    <Sidebar variant="inset" data-testid="wiki-sidebar">
      <motion.div
        className="flex h-full flex-col"
        initial="hidden"
        animate="visible"
        variants={sidebarBlockFade}
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                size="lg"
                className="hover:bg-link-bg-hover"
              >
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-lagoon text-foam shadow-sm shadow-hero-a">
                  <BookOpen className="size-4" />
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-semibold text-sea-ink">
                    {projectName}
                  </span>
                  <span className="text-xs text-sea-ink-soft">
                    Wiki 导航
                  </span>
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
                    {error instanceof Error
                      ? error.message
                      : '加载导航失败'}
                  </div>
                )}

                {!isLoading &&
                  !error &&
                  tree &&
                  tree.root.children &&
                  tree.root.children.length > 0 &&
                  tree.root.children.map((child, i) =>
                    renderNode(child, 0, i),
                  )}

                {!isLoading &&
                  !error &&
                  (!tree ||
                    !tree.root.children ||
                    tree.root.children.length === 0) && (
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
                  由{' '}
                  <span className="font-medium text-sea-ink">
                    Lumina · 微明
                  </span>{' '}
                  驱动
                </span>
              </div>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </motion.div>
    </Sidebar>
  )
}
