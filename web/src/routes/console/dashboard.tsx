import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { useApikeyList } from '#/hooks/useApikey'
import { Plus } from 'lucide-react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'

export const Route = createFileRoute('/console/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data, isLoading } = useApikeyList({ page: 1, size: 1 })

  const totalCount = data?.data?.total_items ?? 0
  const items = data?.data?.items ?? []
  const activeCount = items.filter((item) => item.is_active).length
  const latestCreated = items.length > 0 ? items[0].created_at : null

  return (
    <motion.div
      className="space-y-10"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader title="看板" description="Lumina Console 概览" />

      {/* 欢迎 · 静烛 Hero */}
      <motion.div variants={staggerItem}>
        <div className="flex items-end justify-between gap-6 border-b border-line pb-8">
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.28em] text-lagoon-deep">
              烛照幽微 · 知识中枢
            </p>
            <h2 className="display-title mt-4 text-3xl font-medium text-sea-ink">
              欢迎回来，管理员
            </h2>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-sea-ink-soft">
              这是你的 Lumina 管理面板，在这里管理项目、令牌和系统配置。
            </p>
          </div>
          {!isLoading && totalCount > 0 && (
            <Button
              asChild
              className="shrink-0 bg-sea-ink text-foam hover:bg-lagoon-deep"
            >
              <Link to="/console/apikey">
                <Plus className="size-4" />
                创建新令牌
              </Link>
            </Button>
          )}
        </div>
      </motion.div>

      {/* KPI · 发丝线分栏 */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-3 border-b border-line">
          <div className="border-r border-line py-7 pr-6">
            <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-sea-ink-soft">
              令牌总数
            </p>
            {isLoading ? (
              <Skeleton className="mt-3 h-8 w-16" />
            ) : (
              <p className="display-title mt-3 text-4xl font-medium tracking-tight text-sea-ink">
                {totalCount}
              </p>
            )}
          </div>
          <div className="border-r border-line px-6 py-7">
            <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-sea-ink-soft">
              活跃令牌
            </p>
            {isLoading ? (
              <Skeleton className="mt-3 h-8 w-16" />
            ) : (
              <p className="display-title mt-3 text-4xl font-medium tracking-tight text-sea-ink">
                {activeCount}
              </p>
            )}
          </div>
          <div className="py-7 pl-6">
            <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-sea-ink-soft">
              最近创建
            </p>
            {isLoading ? (
              <Skeleton className="mt-3 h-8 w-28" />
            ) : latestCreated ? (
              <p className="display-title mt-3 text-3xl font-medium tracking-tight text-sea-ink">
                {new Date(latestCreated).toLocaleDateString('zh-CN')}
              </p>
            ) : (
              <p className="mt-4 text-sm text-sea-ink-soft">暂无令牌</p>
            )}
          </div>
        </div>
      </motion.div>

      {/* 快速操作 · 空态 */}
      {!isLoading && totalCount === 0 && (
        <motion.div variants={staggerItem}>
          <div className="border border-dashed border-chip-line py-10 text-center">
            <p className="text-sm text-sea-ink-soft">还没有创建任何 API 令牌</p>
            <Button
              asChild
              className="mt-4 bg-sea-ink text-foam hover:bg-lagoon-deep"
            >
              <Link to="/console/apikey">
                <Plus className="mr-2 size-4" />
                去创建
              </Link>
            </Button>
          </div>
        </motion.div>
      )}
    </motion.div>
  )
}
