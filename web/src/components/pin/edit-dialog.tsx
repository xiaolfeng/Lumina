import { useEffect, useState } from 'react'
import { Button } from '#/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { useUpdatePin } from '#/hooks/usePin'
import type { PinItem } from '#/lib/models/response/pin'

interface EditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  pin: PinItem | null
}

const CATEGORY_OPTIONS = [
  { value: 'notice', label: '注意事项' },
  { value: 'dependency', label: '依赖约束' },
  { value: 'api_change', label: '接口变更' },
  { value: 'other', label: '其他' },
] as const

const PRIORITY_OPTIONS = [
  { value: 'high', label: '高' },
  { value: 'medium', label: '中' },
  { value: 'low', label: '低' },
] as const

export function EditPinDialog({ open, onOpenChange, pin }: EditDialogProps) {
  const [category, setCategory] = useState('notice')
  const [priority, setPriority] = useState('')

  const updateMutation = useUpdatePin()

  useEffect(() => {
    if (pin) {
      setCategory(pin.category || 'notice')
      setPriority(pin.priority)
    }
  }, [pin])

  const handleSubmit = () => {
    if (!pin || !priority) return
    updateMutation.mutate(
      {
        id: pin.id,
        data: {
          category,
          priority,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>编辑 Pin 约束</DialogTitle>
          <DialogDescription>
            修改约束的分类和优先级。标题和内容创建后不可更改。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {/* 标题（只读） */}
          <div className="grid gap-2">
            <Label htmlFor="edit-title">标题</Label>
            <Input id="edit-title" value={pin?.title ?? ''} disabled />
            <p className="text-xs text-muted-foreground">标题创建后不可修改</p>
          </div>

          {/* 内容（只读） */}
          <div className="grid gap-2">
            <Label htmlFor="edit-content">内容</Label>
            <textarea
              id="edit-content"
              value={pin?.content ?? ''}
              disabled
              className="flex min-h-[80px] w-full rounded-md border border-input bg-muted px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-70"
            />
            <p className="text-xs text-muted-foreground">内容创建后不可修改</p>
          </div>

          {/* 分类（可编辑） */}
          <div className="grid gap-2">
            <Label>分类</Label>
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="选择分类" />
              </SelectTrigger>
              <SelectContent>
                {CATEGORY_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* 优先级（可编辑） */}
          <div className="grid gap-2">
            <Label>优先级 *</Label>
            <Select value={priority} onValueChange={setPriority}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="选择优先级" />
              </SelectTrigger>
              <SelectContent>
                {PRIORITY_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!priority || updateMutation.isPending}
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
