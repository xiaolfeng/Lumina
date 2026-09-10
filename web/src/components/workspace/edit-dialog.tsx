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
import { WorkspaceIconPicker } from '#/components/workspace-icon-picker'
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
  const [description, setDescription] = useState('')
  const [icon, setIcon] = useState('')
  const updateMutation = useUpdateWorkspace()
  const pending = updateMutation.isPending

  useEffect(() => {
    if (!open || !item) return
    setName(item.name)
    setDescription(item.description)
    setIcon(item.icon)
  }, [open, item])

  const handleSubmit = () => {
    if (!item || !name.trim() || pending) return
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          name: name.trim(),
          slug: item.slug,
          description: description.trim() || undefined,
          icon: icon.trim() || undefined,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (pending) return
        onOpenChange(next)
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑空间</DialogTitle>
          <DialogDescription>更新空间的名称与图标。</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="ews-name">名称 *</Label>
            <Input
              id="ews-name"
              value={name}
              disabled={pending}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ews-icon">图标</Label>
            <WorkspaceIconPicker
              id="ews-icon"
              value={icon}
              onChange={setIcon}
              disabled={pending}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ews-desc">描述</Label>
            <Input
              id="ews-desc"
              value={description}
              disabled={pending}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            disabled={pending}
            onClick={() => onOpenChange(false)}
          >
            取消
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim() || pending}>
            {pending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
