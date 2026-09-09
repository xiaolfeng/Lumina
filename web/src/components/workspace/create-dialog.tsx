import { useState } from 'react'
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

  const handleSubmit = () => {
    if (!name.trim() || !slug.trim()) return
    createMutation.mutate(
      {
        name: name.trim(),
        slug: slug.trim(),
        description: description.trim() || undefined,
        icon: icon.trim() || undefined,
      },
      { onSuccess: () => handleClose() },
    )
  }

  const handleClose = () => {
    setName('')
    setSlug('')
    setDescription('')
    setIcon('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>新建空间</DialogTitle>
          <DialogDescription>
            用空间把生活与工作项目分开。标识创建后可改，默认空间除外。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="ws-name">名称 *</Label>
            <Input
              id="ws-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：工作"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ws-slug">标识 *</Label>
            <Input
              id="ws-slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="例如：work"
            />
            <p className="text-xs text-muted-foreground">
              小写字母开头，仅字母、数字和连字符，最长 63。
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ws-icon">图标</Label>
            <Input
              id="ws-icon"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
              placeholder="lucide 名如 briefcase，或单个 emoji"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="ws-desc">描述</Label>
            <Input
              id="ws-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="可选"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!name.trim() || !slug.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
