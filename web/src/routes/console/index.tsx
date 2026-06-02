import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Skeleton } from '#/components/ui/skeleton'
import { getApikeyList } from '#/lib/apis/apikey'
import { KeyRound, ShieldCheck, Clock, Plus } from 'lucide-react'

export const Route = createFileRoute('/console/')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['apikey', 'stats'],
    queryFn: () => getApikeyList({ page: 1, size: 1 }),
  })

  const totalCount = data?.data?.total_items ?? 0
  const items = data?.data?.items ?? []
  const activeCount = items.filter((item) => item.is_active).length
  const latestCreated = items.length > 0 ? items[0].created_at : null

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">看板</h1>
        <p className="text-muted-foreground">Lumina Console 概览</p>
      </div>

      {/* KPI Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        {/* 令牌总数 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              令牌总数
            </CardTitle>
            <KeyRound className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <div className="text-2xl font-bold">{totalCount}</div>
            )}
          </CardContent>
        </Card>

        {/* 活跃令牌 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              活跃令牌
            </CardTitle>
            <ShieldCheck className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <div className="text-2xl font-bold">{activeCount}</div>
            )}
          </CardContent>
        </Card>

        {/* 最近创建 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              最近创建
            </CardTitle>
            <Clock className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-32" />
            ) : latestCreated ? (
              <div className="text-lg font-bold">
                {new Date(latestCreated).toLocaleDateString('zh-CN')}
              </div>
            ) : (
              <div className="text-sm text-muted-foreground">暂无令牌</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 快速操作 */}
      <div>
        <h2 className="mb-4 text-lg font-semibold">快速操作</h2>
        {!isLoading && totalCount === 0 ? (
          <Card className="border-dashed">
            <CardContent className="flex flex-col items-center gap-3 py-8">
              <KeyRound className="size-8 text-muted-foreground" />
              <p className="text-center text-muted-foreground">
                还没有创建任何 API 令牌
              </p>
              <Button asChild>
                <Link to="/console/apikey">
                  <Plus className="mr-2 size-4" />
                  去创建
                </Link>
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Button asChild>
            <Link to="/console/apikey">
              <Plus className="mr-2 size-4" />
              创建新令牌
            </Link>
          </Button>
        )}
      </div>
    </div>
  )
}
