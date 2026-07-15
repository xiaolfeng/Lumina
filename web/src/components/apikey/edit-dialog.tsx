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
import { useUpdateApikey } from '#/hooks/useApikey'
import type { ApikeyItem } from '#/lib/models/response/apikey'

interface EditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApikeyItem | null
}

export function EditDialog({ open, onOpenChange, item }: EditDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isActive, setIsActive] = useState(true)

  const updateMutation = useUpdateApikey()

  useEffect(() => {
    if (item) {
      setName(item.name)
      setDescription(item.description || '')
      setIsActive(item.is_active)
    }
  }, [item])

  const handleSubmit = () => {
    if (!item || !name.trim()) return
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          name: name.trim(),
          description: description.trim() || undefined,
          is_active: isActive,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑令牌</DialogTitle>
          <DialogDescription>
            修改令牌的名称、描述和状态。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="edit-name">名称 *</Label>
            <Input
              id="edit-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-desc">描述</Label>
            <Input
              id="edit-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="edit-active">启用状态</Label>
            <button
              id="edit-active"
              type="button"
              role="switch"
              aria-checked={isActive}
              onClick={() => setIsActive(!isActive)}
              className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors ${isActive ? 'bg-primary' : 'bg-input'}`}
            >
              <span
                className={`block size-4 rounded-full bg-background shadow transition-transform ${isActive ? 'translate-x-4' : 'translate-x-0'}`}
              />
            </button>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!name.trim() || updateMutation.isPending}
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
