import type { LucideIcon } from 'lucide-react'
import {
  AlertTriangle,
  Bell,
  BookOpen,
  Box,
  ChevronDown,
  ChevronRight,
  Cloud,
  Code,
  Compass,
  Database,
  File,
  FileText,
  Folder,
  FolderOpen,
  GitBranch,
  Home,
  Info,
  Key,
  Layers,
  Lock,
  Mail,
  Package,
  Search,
  Server,
  Settings,
  Shield,
  Sparkles,
  Terminal,
  Users,
  Workflow,
} from 'lucide-react'
import type { ManifestResponse, WikiNavItem } from './api-client'

/** 运行时内存中的 Wiki 页面节点（与 manifest 对齐，path 无扩展名） */
export interface PageNode {
  path: string
  title: string
  description?: string
  icon?: string
  separator?: string
  defaultOpen?: boolean
  children?: PageNode[]
  /** 父节点引用（由 buildPageTree 建立，用于 sidebar expand state 等） */
  parent?: PageNode
}

/** 运行时内存中的页面树 */
export interface PageTree {
  root: PageNode
  /** 扁平化的有序叶子节点列表（DFS 遍历，跳过 separator 节点） */
  leaves: PageNode[]
}

/** 从 ManifestResponse 递归构建 PageTree，建立 parent 指针并计算 leaves */
export function buildPageTree(manifest: ManifestResponse): PageTree {
  const root: PageNode = {
    path: '',
    title: manifest.meta?.title ?? manifest.project_name,
    description: manifest.meta?.description,
    icon: manifest.meta?.icon,
    children: [],
  }

  const leaves: PageNode[] = []

  function buildChildren(parent: PageNode, items: WikiNavItem[]): void {
    if (items.length === 0) return
    parent.children = items.map((item) => {
      const isSeparator = !!item.separator
      const node: PageNode = {
        path: item.path,
        title: item.title,
        description: item.description,
        icon: item.icon,
        separator: item.separator,
        defaultOpen: item.default_open,
        children: item.children ? [] : undefined,
        parent,
      }

      if (item.children && item.children.length > 0) {
        buildChildren(node, item.children)
      }

      if (!isSeparator && !node.children) {
        leaves.push(node)
      }

      return node
    })
  }

  buildChildren(root, manifest.navigation)

  return { root, leaves }
}

/** 在 PageTree 中按 path 查找节点 */
export function findNode(tree: PageTree, path: string): PageNode | undefined {
  function search(node: PageNode): PageNode | undefined {
    if (node.path === path) return node
    if (node.children) {
      for (const child of node.children) {
        const found = search(child)
        if (found) return found
      }
    }
    return undefined
  }
  return search(tree.root)
}

/** 静态图标映射表：icon name -> LucideIcon 组件 */
const iconMap: Record<string, LucideIcon> = {
  FileText,
  Folder,
  BookOpen,
  Package,
  Compass,
  Code,
  Settings,
  Database,
  Layers,
  Shield,
  Workflow,
  Terminal,
  GitBranch,
  Box,
  Server,
  Cloud,
  Lock,
  Key,
  Users,
  Mail,
  Bell,
  Search,
  Home,
  File,
  FolderOpen,
  ChevronRight,
  ChevronDown,
  Sparkles,
  Info,
  AlertTriangle,
}

/** 根据名称获取 LucideIcon；未知或空值时返回 FileText */
export function getIcon(name?: string): LucideIcon {
  if (!name) return FileText
  return iconMap[name] ?? FileText
}
