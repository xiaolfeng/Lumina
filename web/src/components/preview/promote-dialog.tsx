import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { promotePreviewSession } from '#/lib/apis/pages'
import type { PromoteSessionResponse } from '#/lib/models/response/pages'

export function PromoteDialog({
  open,
  onOpenChange,
  sessionId,
  defaultSlug,
  defaultTitle,
  forked,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  sessionId?: string
  defaultSlug?: string
  defaultTitle?: string
  forked?: boolean
}) {
  const [slug, setSlug] = useState(defaultSlug ?? '')
  const [title, setTitle] = useState(defaultTitle ?? '')
  const [changelog, setChangelog] = useState('')
  const [version, setVersion] = useState('')
  const [setAsActive, setSetAsActive] = useState(true)
  const [pending, setPending] = useState(false)
  const [conflict, setConflict] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    setSlug(defaultSlug ?? '')
    setTitle(defaultTitle ?? '')
    setConflict(null)
  }, [open, defaultSlug, defaultTitle])

  const submit = async (confirmConflict = false) => {
    if (!sessionId) return
    setPending(true)
    try {
      const res = await promotePreviewSession(sessionId, {
        slug,
        title,
        changelog,
        version: version || undefined,
        set_as_active: setAsActive,
        confirm_conflict: confirmConflict,
      })
      const data = res.data as PromoteSessionResponse | undefined
      if (data?.conflict && !confirmConflict) {
        setConflict(data.conflict.message)
        return
      }
      toast.success('已晋升为 Pages')
      onOpenChange(false)
      if (data?.page?.page_url) {
        window.open(data.page.page_url, '_blank')
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '晋升失败')
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border-line sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{forked ? '发布新版本' : '晋升为 Pages'}</DialogTitle>
          <DialogDescription>
            {forked
              ? '将当前草稿发布为新的不可变快照。密码保护请到控制台页面管理设置。'
              : '创建项目级持久页面。密码保护请到控制台页面管理设置。'}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div className="space-y-1">
            <Label>Slug</Label>
            <Input
              value={slug}
              onChange={(event) => setSlug(event.target.value)}
              disabled={forked}
            />
          </div>
          <div className="space-y-1">
            <Label>标题</Label>
            <Input
              value={title}
              onChange={(event) => setTitle(event.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label>版本号（可选）</Label>
            <Input
              value={version}
              placeholder="空则自动递增"
              onChange={(event) => setVersion(event.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label>更新说明</Label>
            <Input
              value={changelog}
              onChange={(event) => setChangelog(event.target.value)}
            />
          </div>
          <label className="flex items-center gap-2 text-sm text-sea-ink">
            <input
              type="checkbox"
              checked={setAsActive}
              onChange={(event) => setSetAsActive(event.target.checked)}
            />
            立即设为当前线上生效版本
          </label>
          {conflict ? (
            <p className="text-sm text-destructive">{conflict}</p>
          ) : null}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            disabled={pending || !title || (!forked && !slug)}
            onClick={() => submit(Boolean(conflict))}
          >
            {conflict ? '确认覆盖并发布' : '发布'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
