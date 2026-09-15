import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { ExternalLink, Shield } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { PageHeader } from '#/components/page-header'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import {
  useArchivePage,
  usePageList,
  useUpdateAccessPolicy,
} from '#/hooks/usePages'
import { formatDateTime } from '#/lib/format-date'
import type { PageItem } from '#/lib/models/response/pages'

export const Route = createFileRoute('/console/pages/')({
  staticData: { crumb: '页面' },
  component: ConsolePagesPage,
})

function ConsolePagesPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [archiveTarget, setArchiveTarget] = useState<PageItem | null>(null)
  const [policyTarget, setPolicyTarget] = useState<PageItem | null>(null)
  const [accessMode, setAccessMode] = useState<'public' | 'password'>('public')
  const [password, setPassword] = useState('')
  const { current } = useCurrentWorkspace()
  const { data, isLoading } = usePageList({
    page,
    size: pageSize,
    workspace_id: current?.id,
  })
  const archive = useArchivePage()
  const updatePolicy = useUpdateAccessPolicy()
  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  return (
    <div className="mx-auto max-w-5xl space-y-8">
      <PageHeader
        title="页面管理"
        description="管理已发布的即时页面、版本生效指针与访问策略。密码保护只在这里设置。"
      />
      <div className="border-t border-line">
        {isLoading ? (
          <p className="py-8 text-sm text-sea-ink-soft">加载中…</p>
        ) : items.length === 0 ? (
          <div className="px-8 py-16 text-center">
            <p className="display-title text-xl font-semibold text-sea-ink">
              暂无页面
            </p>
            <p className="mt-2 text-sm text-sea-ink-soft">
              从 Preview 工作台晋升后会出现在这里。
            </p>
          </div>
        ) : (
          items.map((item) => (
            <div
              key={item.id}
              className="flex items-center gap-4 border-b border-line px-1 py-4"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-sea-ink">
                  {item.title}
                </p>
                <p className="mt-1 text-xs text-sea-ink-soft">
                  {item.project_name}/{item.slug} · {item.latest_version} ·{' '}
                  {item.access_mode === 'password' ? '密码保护' : '公开'} ·{' '}
                  {formatDateTime(item.updated_at)}
                </p>
              </div>
              {/* page_url 的域名在 site.domain 未配置时回退 localhost，跨部署不可靠；
                  改用相对裸路径，入口由展示页按 version.entry_filename 兜底解析 */}
              <a
                href={`/pages/${item.project_name}/${item.slug}/`}
                target="_blank"
                rel="noopener noreferrer"
                className="grid size-11 place-items-center text-sea-ink-soft"
              >
                <ExternalLink className="size-3.5" />
              </a>
              <button
                type="button"
                className="grid size-11 place-items-center text-sea-ink-soft"
                onClick={() => {
                  setPolicyTarget(item)
                  setAccessMode(item.access_mode)
                  setPassword('')
                }}
              >
                <Shield className="size-3.5" />
              </button>
              {item.status === 'published' ? (
                <Button
                  variant="ghost"
                  onClick={() => setArchiveTarget(item)}
                >
                  归档
                </Button>
              ) : (
                <span className="text-xs text-sea-ink-soft">已归档</span>
              )}
            </div>
          ))
        )}
      </div>
      {totalItems > 0 ? (
        <DataTablePagination
          currentPage={page}
          totalPages={totalPages}
          totalItems={totalItems}
          pageSize={pageSize}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      ) : null}

      <ConfirmDeleteDialog
        open={!!archiveTarget}
        onOpenChange={(open) => {
          if (!open) setArchiveTarget(null)
        }}
        title="归档页面"
        description={`确定归档「${archiveTarget?.title}」？归档后对外路径将不可访问。`}
        onConfirm={() => {
          if (archiveTarget) archive.mutate(archiveTarget.id)
          setArchiveTarget(null)
        }}
        isPending={archive.isPending}
      />

      <Dialog
        open={!!policyTarget}
        onOpenChange={(open) => {
          if (!open) setPolicyTarget(null)
        }}
      >
        <DialogContent className="border-line sm:max-w-md">
          <DialogHeader>
            <DialogTitle>页面安全设置</DialogTitle>
            <DialogDescription>
              访问策略只在管理员端口配置，不会出现在晋升弹窗里。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={accessMode === 'public'}
                onChange={() => setAccessMode('public')}
              />
              公开访问
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={accessMode === 'password'}
                onChange={() => setAccessMode('password')}
              />
              密码保护
            </label>
            {accessMode === 'password' ? (
              <div className="space-y-1">
                <Label>访问密码</Label>
                <Input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
              </div>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPolicyTarget(null)}>
              取消
            </Button>
            <Button
              disabled={updatePolicy.isPending || (accessMode === 'password' && !password)}
              onClick={() => {
                if (!policyTarget) return
                updatePolicy.mutate({
                  id: policyTarget.id,
                  access_mode: accessMode,
                  password: accessMode === 'password' ? password : undefined,
                })
                setPolicyTarget(null)
              }}
            >
              保存生效
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
