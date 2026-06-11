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
import { useUpdateProject } from '#/hooks/useProject'
import type { ProjectItem } from '#/lib/models/response/project'

interface EditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ProjectItem | null
}

export function EditDialog({ open, onOpenChange, item }: EditDialogProps) {
  const [name, setName] = useState('')
  const [aliasName, setAliasName] = useState('')
  const [matchPathInput, setMatchPathInput] = useState('')
  const [description, setDescription] = useState('')

  const updateMutation = useUpdateProject()

  useEffect(() => {
    if (item) {
      setName(item.name)
      setAliasName(item.alias_name || '')
      setMatchPathInput(item.match_path?.join(', ') ?? '')
      setDescription(item.description ?? '')
    }
  }, [item])

  const handleSubmit = () => {
    if (!item || !name.trim()) return
    const matchPaths = matchPathInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          name: name.trim(),
          alias_name: aliasName.trim() || undefined,
          match_path: matchPaths,
          description: description.trim() || undefined,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  const handleClose = () => {
    setName('')
    setAliasName('')
    setMatchPathInput('')
    setDescription('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑项目</DialogTitle>
          <DialogDescription>修改项目的基本信息。</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="e-name">项目名称 *</Label>
            <Input
              id="e-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="e-alias">别名</Label>
            <Input
              id="e-alias"
              value={aliasName}
              onChange={(e) => setAliasName(e.target.value)}
              placeholder="输入项目别名"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="e-match-path">匹配路径</Label>
            <Input
              id="e-match-path"
              value={matchPathInput}
              onChange={(e) => setMatchPathInput(e.target.value)}
              placeholder="逗号分隔，如: /api/v1/*,/docs/*"
            />
            <p className="text-xs text-muted-foreground">
              支持通配符 * 匹配，用于自动关联请求路径
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="e-desc">描述</Label>
            <textarea
              id="e-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
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
