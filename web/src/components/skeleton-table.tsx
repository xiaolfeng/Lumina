import { Skeleton } from '@lumina/components/ui/skeleton'

interface SkeletonTableProps {
  /** 渲染行数，默认 3 */
  rows?: number
}

/**
 * 表格加载骨架屏
 *
 * 用于列表页 isLoading 状态下展示占位行，
 * 保持页面高度稳定，避免内容跳动。
 */
export function SkeletonTable({ rows = 3 }: SkeletonTableProps) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <Skeleton key={i} className="h-8 w-full" />
      ))}
    </div>
  )
}
