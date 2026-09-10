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
import { useCreateWorkspace } from '#/hooks/useWorkspace'
import { WorkspaceIconPicker } from '#/components/workspace-icon-picker'
import { createHiddenWorkspaceSlug } from '#/components/workspace-icon-utils'

interface CreateWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
}: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [icon, setIcon] = useState('')
  const createMutation = useCreateWorkspace()
  const pending = createMutation.isPending

  useEffect(() => {
    if (!open) return
    setName('')
    setSlug(createHiddenWorkspaceSlug())
    setDescription('')
    setIcon('')
  }, [open])

  const handleSubmit = () => {
    if (!name.trim() || !slug || pending) return
    createMutation.mutate(
      {
        name: name.trim(),
        slug,
        description: description.trim() || undefined,
        icon: icon.trim() || undefined,
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
          <DialogTitle>新建空间</DialogTitle>
          <DialogDescription>用空间把生活与工作项目分开。</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="ws-name">名称 *</Label>
            <Input
              id="ws-name"
              value={name}
              disabled={pending}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：工作"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ws-icon">图标</Label>
            <WorkspaceIconPicker
              id="ws-icon"
              value={icon}
              onChange={setIcon}
              disabled={pending}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ws-desc">描述</Label>
            <Input
              id="ws-desc"
              value={description}
              disabled={pending}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="可选"
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
            {pending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
