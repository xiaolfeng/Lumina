import { useEffect, useState } from 'react'
import {
  Check,
  Copy,
  ExternalLink,
  GitFork,
  History,
  Lock,
  Shield,
  Unlock,
} from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import {
  useForkPage,
  usePageVersions,
  useSwitchActiveVersion,
  useUpdateAccessPolicy,
} from '#/hooks/usePages'
import { formatDateTime } from '#/lib/format-date'
import type { PageItem, PageVersionItem } from '#/lib/models/response/pages'
import { toast } from 'sonner'

interface PageDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: PageItem | null
  initialMode?: 'view' | 'policy'
}

export function PageDetailSheet({
  open,
  onOpenChange,
  item,
  initialMode = 'view',
}: PageDetailSheetProps) {
  const [copiedSlug, setCopiedSlug] = useState(false)
  const [currentItem, setCurrentItem] = useState<PageItem | null>(item)
  const [isEditingPolicy, setIsEditingPolicy] = useState(
    initialMode === 'policy',
  )
  const [accessMode, setAccessMode] = useState<'public' | 'password'>('public')
  const [password, setPassword] = useState('')

  const { data: versionsData, isLoading: versionsLoading } = usePageVersions(
    currentItem?.id,
  )
  const versions: PageVersionItem[] = versionsData?.data?.items ?? []

  const updatePolicy = useUpdateAccessPolicy()
  const switchVersion = useSwitchActiveVersion()
  const forkMutation = useForkPage()

  useEffect(() => {
    if (open && item) {
      setCurrentItem(item)
      setIsEditingPolicy(initialMode === 'policy')
      setAccessMode(item.access_mode)
      setPassword('')
      setCopiedSlug(false)
    }
  }, [open, item, initialMode])

  if (!currentItem) return null

  const isPasswordProtected = currentItem.access_mode === 'password'
  const isPublished = currentItem.status === 'published'
  const publicUrl = `/pages/${currentItem.project_name || currentItem.project_id}/${currentItem.slug}`

  const handleCopyUrl = () => {
    const full = `${window.location.origin}${publicUrl}`
    void navigator.clipboard.writeText(full)
    setCopiedSlug(true)
    setTimeout(() => setCopiedSlug(false), 2000)
    toast.success('页面公开访问地址已复制')
  }

  const handleSavePolicy = () => {
    if (accessMode === 'password' && !password.trim()) {
      toast.error('表单校验未通过', {
        description: '启用密码门时必须输入有效口令',
      })
      return
    }
    updatePolicy.mutate(
      {
        id: currentItem.id,
        access_mode: accessMode,
        password: accessMode === 'password' ? password : undefined,
      },
      {
        onSuccess: () => {
          setIsEditingPolicy(false)
          setPassword('')
          setCurrentItem((prev) =>
            prev ? { ...prev, access_mode: accessMode } : null,
          )
        },
      },
    )
  }

  const handleSwitchVersion = (versionId: string) => {
    switchVersion.mutate(
      {
        id: currentItem.id,
        versionId,
      },
      {
        onSuccess: () => {
          setCurrentItem((prev) =>
            prev ? { ...prev, latest_version_id: versionId } : null,
          )
        },
      },
    )
  }

  const handleFork = (versionId?: string) => {
    forkMutation.mutate(
      { id: currentItem.id, versionId },
      {
        onSuccess: (res) => {
          toast.success('已从该快照派生新 Preview 沙盒会话')
          const d = res.data
          const url =
            d?.preview_url ||
            (d?.session.hash ? `/preview/${d.session.hash}/index.html` : null)
          if (url) {
            window.open(url, '_blank')
          }
        },
      },
    )
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="w-full rounded-none border-l border-line bg-foam p-0 sm:max-w-xl flex flex-col"
      >
        <SheetHeader className="border-b border-line bg-surface-strong px-6 py-5 text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="size-2 rotate-45 bg-lagoon" />
              <SheetTitle className="text-base font-semibold text-sea-ink">
                {isEditingPolicy ? '配置访问策略' : '即时页面全景资产'}
              </SheetTitle>
            </div>
            <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
              {isEditingPolicy ? 'POLICY' : 'PAGE'}
            </span>
          </div>
          <SheetDescription className="text-xs text-sea-ink-soft">
            {isEditingPolicy
              ? '配置不可变页面快照的对外访问权限与密码门保护'
              : '管理生效版本指针、访问策略、查看快照版本时间线与 Fork 派生'}
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5 text-[13px]">
          {isEditingPolicy ? (
            /* ── 策略编辑模式 ── */
            <div className="space-y-4">
              <div className="border border-line bg-sand/40 p-3 space-y-2 text-xs">
                <span className="font-semibold text-sea-ink block">
                  密码门访问策略说明：
                </span>
                <p className="text-sea-ink-soft leading-relaxed">
                  密码门策略通过安全 HMAC Cookie
                  签发会话凭据。启用密码保护后，外部访客需在门锁页输入口令方可访问。
                </p>
              </div>

              <div className="space-y-2">
                <Label className="text-xs font-semibold text-sea-ink">
                  访问安全模式 <span className="text-destructive">*</span>
                </Label>
                <div className="grid grid-cols-2 gap-3">
                  <div
                    onClick={() => setAccessMode('public')}
                    className={`cursor-pointer border p-3 flex flex-col gap-1 transition-colors ${
                      accessMode === 'public'
                        ? 'border-lagoon bg-foam shadow-sm'
                        : 'border-line bg-surface hover:bg-sand/60'
                    }`}
                  >
                    <div className="flex items-center gap-1.5 font-semibold text-xs text-sea-ink">
                      <Unlock className="size-3.5 text-lagoon" />
                      <span>公开直出</span>
                    </div>
                    <span className="text-[11px] text-sea-ink-soft">
                      任何访客均可自由直接访问
                    </span>
                  </div>

                  <div
                    onClick={() => setAccessMode('password')}
                    className={`cursor-pointer border p-3 flex flex-col gap-1 transition-colors ${
                      accessMode === 'password'
                        ? 'border-lagoon bg-foam shadow-sm'
                        : 'border-line bg-surface hover:bg-sand/60'
                    }`}
                  >
                    <div className="flex items-center gap-1.5 font-semibold text-xs text-sea-ink">
                      <Lock className="size-3.5 text-lagoon" />
                      <span>密码保护门</span>
                    </div>
                    <span className="text-[11px] text-sea-ink-soft">
                      需输入指定访问口令解锁
                    </span>
                  </div>
                </div>
              </div>

              {accessMode === 'password' && (
                <div className="space-y-1.5">
                  <Label className="text-xs font-semibold text-sea-ink">
                    设置访问口令 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="输入新的安全访问口令..."
                    className="bg-sand/40 text-xs font-mono"
                  />
                  <p className="text-[11px] text-sea-ink-soft">
                    保存后立即生效并使此前签发的过期会话凭据失效。
                  </p>
                </div>
              )}
            </div>
          ) : (
            /* ── 查看全景模式 ── */
            <>
              {/* Slug 标识与状态 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <div className="flex items-center justify-between">
                  <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                    页面 Slug 标识与状态
                  </span>
                  <div className="flex items-center gap-1.5">
                    <span
                      className={`font-mono text-[10px] px-1.5 py-0.5 border ${
                        isPublished
                          ? 'border-chip-line bg-lagoon/10 text-lagoon-deep font-semibold'
                          : 'border-line bg-chip-bg text-sea-ink-soft'
                      }`}
                    >
                      {isPublished ? '已上线' : '已归档'}
                    </span>
                    <span
                      className={`font-mono text-[10px] px-1.5 py-0.5 border ${
                        isPasswordProtected
                          ? 'border-destructive/30 bg-destructive/10 text-destructive'
                          : 'border-line bg-sand text-sea-ink-soft'
                      }`}
                    >
                      {isPasswordProtected ? '密码保护门' : '公开直出'}
                    </span>
                  </div>
                </div>
                <div className="flex items-center justify-between bg-sand border border-line p-2 mt-1">
                  <span className="font-mono text-xs text-sea-ink select-all">
                    {publicUrl}
                  </span>
                  <button
                    type="button"
                    onClick={handleCopyUrl}
                    className="inline-flex items-center gap-1 text-[11px] text-sea-ink-soft hover:text-lagoon transition-colors"
                    title="复制访问链接"
                  >
                    {copiedSlug ? (
                      <>
                        <Check className="size-3 text-lagoon" />
                        <span>已复制</span>
                      </>
                    ) : (
                      <>
                        <Copy className="size-3" />
                        <span>复制</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              {/* 归属项目与访问策略 */}
              <div className="grid grid-cols-2 gap-3 border-b border-dashed border-line pb-4">
                <div className="bg-sand border border-line p-2.5">
                  <span className="text-[10px] text-sea-ink-soft uppercase font-bold block mb-1">
                    所属项目
                  </span>
                  <span className="font-semibold text-xs text-sea-ink truncate block">
                    {currentItem.project_name || '未关联项目'}
                  </span>
                </div>
                <div className="bg-sand border border-line p-2.5">
                  <span className="text-[10px] text-sea-ink-soft uppercase font-bold block mb-1">
                    线上版本 ID
                  </span>
                  <span className="font-mono text-xs text-lagoon-deep font-semibold truncate block">
                    {currentItem.latest_version_id
                      ? currentItem.latest_version_id.slice(0, 12)
                      : '无版本'}
                  </span>
                </div>
              </div>

              {/* 版本快照历史时间线 */}
              <div className="space-y-3 border-b border-dashed border-line pb-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <History className="size-3.5 text-lagoon" />
                    <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                      不可变版本快照历史 ({versions.length})
                    </span>
                  </div>
                  <span className="text-[11px] text-sea-ink-soft">
                    支持一键切换生效指针
                  </span>
                </div>

                {versionsLoading ? (
                  <div className="text-center py-6 text-xs text-sea-ink-soft">
                    加载版本时间线中...
                  </div>
                ) : versions.length === 0 ? (
                  <div className="border border-dashed border-line p-4 text-center text-xs text-sea-ink-soft">
                    暂无历史版本快照
                  </div>
                ) : (
                  <div className="space-y-2 max-h-56 overflow-y-auto pr-1">
                    {versions.map((ver) => {
                      const isCurrentActive =
                        currentItem.latest_version_id === ver.id
                      return (
                        <div
                          key={ver.id}
                          className={`p-3 border flex flex-col gap-1.5 transition-colors ${
                            isCurrentActive
                              ? 'border-lagoon bg-sand/50 shadow-sm'
                              : 'border-line bg-surface hover:bg-sand/30'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span className="font-mono text-xs font-semibold text-sea-ink">
                                {ver.version}
                              </span>
                              {isCurrentActive && (
                                <span className="font-mono text-[9.5px] px-1.5 py-0.2 bg-lagoon text-foam font-bold">
                                  CURRENT ACTIVE
                                </span>
                              )}
                            </div>
                            <span className="font-mono text-[10px] text-sea-ink-soft">
                              {formatDateTime(ver.created_at)}
                            </span>
                          </div>

                          {ver.changelog && (
                            <p className="text-xs text-sea-ink-soft line-clamp-1">
                              {ver.changelog}
                            </p>
                          )}

                          <div className="flex items-center justify-between pt-1 border-t border-dashed border-line/60">
                            <span className="font-mono text-[10px] text-sea-ink-soft">
                              {ver.file_count} 个静态文件
                            </span>
                            <div className="flex items-center gap-2">
                              <button
                                type="button"
                                onClick={() => handleFork(ver.id)}
                                className="inline-flex items-center gap-1 text-[11px] text-sea-ink-soft hover:text-lagoon transition-colors"
                                title="从该快照派生 Preview 沙盒"
                              >
                                <GitFork className="size-3" />
                                <span>Fork</span>
                              </button>
                              {!isCurrentActive && (
                                <button
                                  type="button"
                                  onClick={() => handleSwitchVersion(ver.id)}
                                  disabled={switchVersion.isPending}
                                  className="text-[11px] text-lagoon-deep font-semibold hover:underline"
                                >
                                  设为生效
                                </button>
                              )}
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>

              {/* 审计时间 */}
              <div className="space-y-2 text-xs text-sea-ink-soft">
                <div className="flex justify-between py-1 border-b border-line">
                  <span>实体 ID</span>
                  <span className="font-mono text-sea-ink">
                    {currentItem.id}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b border-line">
                  <span>创建时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(currentItem.created_at)}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b border-line">
                  <span>更新时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(currentItem.updated_at)}
                  </span>
                </div>
              </div>
            </>
          )}
        </div>

        {/* 底部操作工具栏 */}
        <div className="border-t border-line bg-surface-strong px-6 py-4 flex items-center justify-between gap-3">
          {isEditingPolicy ? (
            <>
              <Button
                type="button"
                variant="outline"
                className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
                onClick={() => setIsEditingPolicy(false)}
                disabled={updatePolicy.isPending}
              >
                取消
              </Button>
              <Button
                type="button"
                className="rounded-none bg-primary text-primary-foreground hover:bg-primary/90 flex-1"
                onClick={handleSavePolicy}
                disabled={
                  updatePolicy.isPending ||
                  (accessMode === 'password' && !password.trim())
                }
              >
                {updatePolicy.isPending ? '保存中...' : '保存策略'}
              </Button>
            </>
          ) : (
            <>
              <Button
                type="button"
                variant="outline"
                className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
                onClick={() => setIsEditingPolicy(true)}
              >
                <Shield className="mr-1.5 size-3.5 text-lagoon" />
                调整访问策略
              </Button>
              {isPublished ? (
                <Button
                  type="button"
                  className="rounded-none bg-lagoon text-foam hover:bg-lagoon-deep flex-1"
                  onClick={() => window.open(publicUrl, '_blank')}
                >
                  <span>直出访问页面</span>
                  <ExternalLink className="ml-1.5 size-3.5" />
                </Button>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  disabled
                  className="rounded-none border-line text-sea-ink-soft/60 flex-1 cursor-not-allowed"
                >
                  <span>页面已归档下线</span>
                </Button>
              )}
            </>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
