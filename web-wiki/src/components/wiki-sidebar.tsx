/**
 * Wiki 侧边栏导航组件
 *
 * 功能特性：
 * - 从 manifest API 获取导航结构（文件树）
 * - 支持目录展开/折叠交互
 * - 当前页面路径高亮显示
 * - 使用 TanStack Router Link 进行客户端导航
 * - 响应式设计：移动端可隐藏
 */
import { useState, useEffect } from 'react'
import { Link } from '@tanstack/react-router'
import {
  ChevronRight,
  ChevronDown,
  FolderOpen,
  FolderClosed,
  FileText,
  BookOpen,
  Loader2,
} from 'lucide-react'
import { wikiReaderApi } from '#/lib/api-client'
import type { ManifestResponse } from '#/lib/api-client'

/** 导航条目类型（与后端 WikiNavItem 对齐） */
interface NavEntry {
  path: string
  title: string
  children?: NavEntry[]
}

interface WikiSidebarProps {
  wikiId: string
  currentPagePath?: string
  isOpen: boolean
  onToggle: () => void
}

export function WikiSidebar({
  wikiId,
  currentPagePath = '',
  isOpen,
  onToggle,
}: WikiSidebarProps) {
  const [manifest, setManifest] = useState<ManifestResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set())

  // 获取 manifest 数据
  useEffect(() => {
    const fetchManifest = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await wikiReaderApi.getManifest(wikiId)
        setManifest(data)

        // 默认展开所有包含当前页面的父目录
        if (currentPagePath) {
          const dirsToExpand = new Set<string>()
          const pathParts = currentPagePath.split('/').filter(Boolean)
          let currentPath = ''
          for (const part of pathParts.slice(0, -1)) {
            currentPath += (currentPath ? '/' : '') + part
            dirsToExpand.add(currentPath)
          }
          setExpandedDirs(dirsToExpand)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : '加载导航失败')
      } finally {
        setLoading(false)
      }
    }

    fetchManifest()
  }, [wikiId])

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

  // 渲染单个导航项
  const renderNavItem = (entry: NavEntry, depth: number = 0) => {
    const isExpanded = expandedDirs.has(entry.path)
    const isDirectory = entry.children !== undefined
    const isActive = !isDirectory && entry.path === currentPagePath

    return (
      <div key={entry.path} className="nav-item">
        <div
          className={`group flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent ${
            depth > 0 ? 'ml-' + Math.min(depth * 3, 9) : ''
          } ${isActive ? 'bg-accent text-lagoon font-medium' : 'text-sea-ink-soft'}`}
          style={{ paddingLeft: `${depth * 12 + 8}px` }}
        >
          {isDirectory ? (
            <>
              <button
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  toggleDir(entry.path)
                }}
                className="inline-flex h-4 w-4 items-center justify-center rounded hover:bg-muted"
              >
                {isExpanded ? (
                  <ChevronDown className="h-3 w-3" />
                ) : (
                  <ChevronRight className="h-3 w-3" />
                )}
              </button>
              <Link
                to="/wiki/$wikiId/$"
                params={{ wikiId, _splat: entry.path }}
                className="flex flex-1 items-center gap-2"
              >
                {isExpanded ? (
                  <FolderOpen className="h-4 w-4 text-lagoon" />
                ) : (
                  <FolderClosed className="h-4 w-4 text-muted-foreground" />
                )}
                <span className="truncate">{entry.title}</span>
              </Link>
            </>
          ) : (
            <>
              <span className="w-4" /> {/* 占位，保持对齐 */}
              <Link
                to="/wiki/$wikiId/$"
                params={{ wikiId, _splat: entry.path }}
                className={`flex flex-1 items-center gap-2 ${isActive ? 'text-lagoon' : ''}`}
              >
                <FileText className="h-4 w-4" />
                <span className="truncate">{entry.title}</span>
              </Link>
            </>
          )}
        </div>

        {/* 子目录递归渲染 */}
        {isDirectory && isExpanded && entry.children && (
          <div className="children">
            {entry.children.map((child) => renderNavItem(child, depth + 1))}
          </div>
        )}
      </div>
    )
  }

  return (
    <>
      {/* 移动端遮罩层 */}
      {isOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/20 backdrop-blur-sm md:hidden"
          onClick={onToggle}
        />
      )}

      {/* 侧边栏主体 */}
      <aside
        className={`sidebar fixed left-0 top-0 z-40 flex h-full w-72 flex-col border-r border-line bg-surface-strong backdrop-blur-xl transition-transform duration-300 ease-in-out md:relative md:z-auto md:translate-x-0 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* 头部 */}
        <div className="flex items-center justify-between border-b border-line px-4 py-3">
          <div className="flex items-center gap-2">
            <BookOpen className="h-5 w-5 text-lagoon" />
            <span className="font-semibold text-sea-ink">Wiki 导航</span>
          </div>
          {/* 关闭按钮（移动端） */}
          <button
            onClick={onToggle}
            className="rounded-md p-1 hover:bg-accent md:hidden"
            aria-label="关闭侧边栏"
          >
            ✕
          </button>
        </div>

        {/* 导航内容区 */}
        <nav className="flex-1 overflow-y-auto p-2">
          {/* 首页链接 */}
          <Link
            to="/wiki/$wikiId"
            params={{ wikiId }}
            className={`mb-2 flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium transition-colors hover:bg-accent ${
              !currentPagePath ? 'bg-accent text-lagoon' : 'text-sea-ink-soft'
            }`}
          >
            <BookOpen className="h-4 w-4" />
            <span>首页</span>
          </Link>

          {/* 加载状态 */}
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-lagoon" />
              <span className="ml-2 text-sm text-sea-ink-soft">加载中...</span>
            </div>
          )}

          {/* 错误提示 */}
          {error && (
            <div className="mx-2 rounded-md bg-destructive/10 p-3 text-xs text-destructive">
              {error}
            </div>
          )}

          {/* 导航树 */}
          {!loading && !error && manifest && manifest.navigation.length > 0 && (
            <div className="nav-tree space-y-0.5">
              {manifest.navigation.map((entry) => renderNavItem(entry))}
            </div>
          )}

          {/* 空状态 */}
          {!loading &&
            !error &&
            (!manifest?.navigation || manifest.navigation.length === 0) && (
              <div className="py-8 text-center text-sm text-sea-ink-soft">
                暂无页面
              </div>
            )}
        </nav>

        {/* 底部信息 */}
        <div className="border-t border-line px-4 py-2 text-xs text-muted-foreground">
          Wiki Reader v0.1.0
        </div>
      </aside>
    </>
  )
}
