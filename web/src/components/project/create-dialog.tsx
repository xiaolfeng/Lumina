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
import { useCreateProject } from '#/hooks/useProject'

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateDialog({ open, onOpenChange }: CreateDialogProps) {
  const [name, setName] = useState('')
  const [aliasName, setAliasName] = useState('')
  const [matchPathInput, setMatchPathInput] = useState('')
  const [description, setDescription] = useState('')

  const createMutation = useCreateProject()

  const handleSubmit = () => {
    if (!name.trim()) return
    const matchPaths = matchPathInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    createMutation.mutate(
      {
        name: name.trim(),
        alias_name: aliasName.trim() || undefined,
        match_path: matchPaths.length > 0 ? matchPaths : undefined,
        description: description.trim() || undefined,
      },
      {
        onSuccess: () => handleClose(),
      },
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
          <DialogTitle>创建项目</DialogTitle>
          <DialogDescription>
            创建一个新的项目，用于组织 Pin 和 Q&A。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="p-name">项目名称 *</Label>
            <Input
              id="p-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="输入项目名称"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="p-alias">别名</Label>
            <Input
              id="p-alias"
              value={aliasName}
              onChange={(e) => setAliasName(e.target.value)}
              placeholder="输入项目别名"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="p-match-path">匹配路径</Label>
            <Input
              id="p-match-path"
              value={matchPathInput}
              onChange={(e) => setMatchPathInput(e.target.value)}
              placeholder="逗号分隔，如: /api/v1/*,/docs/*"
            />
            <p className="text-xs text-muted-foreground">
              支持通配符 * 匹配，用于自动关联请求路径
            </p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="p-desc">描述</Label>
            <textarea
              id="p-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="输入项目描述（可选）"
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
            disabled={!name.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
