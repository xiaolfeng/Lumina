import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { useApikeyList } from '#/hooks/useApikey'
import { KeyRound, ShieldCheck, Clock, Plus, Sparkles } from 'lucide-react'
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
      className="space-y-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader title="看板" description="Lumina Console 概览" />

      {/* 欢迎横幅 */}
      <motion.div variants={staggerItem}>
        <Card className="border-chip-line bg-gradient-to-r from-surface-strong to-surface overflow-hidden">
          <CardContent className="flex items-center gap-4 py-4">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-lagoon/15 text-lagoon">
              <Sparkles className="size-5" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-medium text-sea-ink">
                欢迎回来，管理员
              </p>
              <p className="text-xs text-sea-ink-soft">
                这是你的 Lumina 管理面板，在这里管理项目、令牌和系统配置
              </p>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* KPI Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        {/* 令牌总数 */}
        <motion.div variants={staggerItem}>
          <Card className="transition-shadow duration-200 hover:shadow-md hover:shadow-hero-a">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-sea-ink-soft">
                令牌总数
              </CardTitle>
              <div className="flex size-8 items-center justify-center rounded-lg bg-lagoon/10 text-lagoon">
                <KeyRound className="size-4" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                <div className="text-2xl font-bold text-sea-ink">{totalCount}</div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* 活跃令牌 */}
        <motion.div variants={staggerItem}>
          <Card className="transition-shadow duration-200 hover:shadow-md hover:shadow-hero-b">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-sea-ink-soft">
                活跃令牌
              </CardTitle>
              <div className="flex size-8 items-center justify-center rounded-lg bg-palm/10 text-palm">
                <ShieldCheck className="size-4" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                <div className="text-2xl font-bold text-sea-ink">{activeCount}</div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* 最近创建 */}
        <motion.div variants={staggerItem}>
          <Card className="transition-shadow duration-200 hover:shadow-md hover:shadow-hero-a">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-sea-ink-soft">
                最近创建
              </CardTitle>
              <div className="flex size-8 items-center justify-center rounded-lg bg-kicker/10 text-kicker">
                <Clock className="size-4" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-8 w-32" />
              ) : latestCreated ? (
                <div className="text-lg font-bold text-sea-ink">
                  {new Date(latestCreated).toLocaleDateString('zh-CN')}
                </div>
              ) : (
                <div className="text-sm text-sea-ink-soft">暂无令牌</div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      </div>

      {/* 快速操作 */}
      <motion.div variants={staggerItem}>
        <h2 className="mb-4 text-lg font-semibold text-sea-ink">快速操作</h2>
        {!isLoading && totalCount === 0 ? (
          <Card className="border-dashed border-chip-line">
            <CardContent className="flex flex-col items-center gap-3 py-8">
              <div className="flex size-12 items-center justify-center rounded-xl bg-lagoon/10 text-lagoon">
                <KeyRound className="size-6" />
              </div>
              <p className="text-center text-sea-ink-soft">
                还没有创建任何 API 令牌
              </p>
              <Button asChild className="bg-lagoon text-foam hover:bg-lagoon-deep">
                <Link to="/console/apikey">
                  <Plus className="mr-2 size-4" />
                  去创建
                </Link>
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Button asChild className="bg-lagoon text-foam hover:bg-lagoon-deep">
            <Link to="/console/apikey">
              <Plus className="mr-2 size-4" />
              创建新令牌
            </Link>
          </Button>
        )}
      </motion.div>
    </motion.div>
  )
}
