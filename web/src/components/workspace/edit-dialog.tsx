import { useEffect, useState } from 'react'
import { Button } from '@lumina/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { useUpdateWorkspace } from '#/hooks/useWorkspace'
import type { WorkspaceItem } from '#/lib/models/response/workspace'

interface EditWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: WorkspaceItem | null
}

export function EditWorkspaceDialog({
  open,
  onOpenChange,
  item,
}: EditWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [icon, setIcon] = useState('')
  const updateMutation = useUpdateWorkspace()
  const slugLocked = Boolean(item?.is_default)

  useEffect(() => {
    if (!item) return
    setName(item.name)
    setSlug(item.slug)
    setDescription(item.description ?? '')
    setIcon(item.icon ?? '')
  }, [item])

  const handleSubmit = () => {
    if (!item || !name.trim() || !slug.trim()) return
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          name: name.trim(),
          slug: slug.trim(),
          description: description.trim() || undefined,
          icon: icon.trim() || undefined,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑空间</DialogTitle>
          <DialogDescription>
            {slugLocked
              ? '默认空间可以改名称，标识不可修改。'
              : '更新空间名称、标识与图标。'}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="ews-name">名称 *</Label>
            <Input
              id="ews-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ews-slug">标识 *</Label>
            <Input
              id="ews-slug"
              value={slug}
              disabled={slugLocked}
              onChange={(e) => setSlug(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ews-icon">图标</Label>
            <Input
              id="ews-icon"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ews-desc">描述</Label>
            <Input
              id="ews-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!name.trim() || !slug.trim() || updateMutation.isPending}
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
