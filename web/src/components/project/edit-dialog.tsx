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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@lumina/components/ui/select'
import { WorkspaceIcon } from '#/components/workspace-icon'
import { useWorkspaceOptions } from '#/hooks/useWorkspace'
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
  const [workspaceId, setWorkspaceId] = useState('')
  const workspaceOptions = useWorkspaceOptions()
  const workspaces = workspaceOptions.data ?? []

  const updateMutation = useUpdateProject()
  const targetExists = workspaces.some(
    (workspace) => workspace.id === workspaceId,
  )
  const canSubmit =
    Boolean(item && name.trim() && targetExists) &&
    !workspaceOptions.isError &&
    !updateMutation.isPending

  useEffect(() => {
    if (open && item) {
      setName(item.name)
      setAliasName(item.alias_name || '')
      setMatchPathInput(item.match_path?.join(', ') ?? '')
      setDescription(item.description)
      setWorkspaceId(item.workspace_id)
    }
  }, [item, open])

  const handleSubmit = () => {
    if (!item || !canSubmit) return
    const matchPaths = matchPathInput
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          workspace_id:
            workspaceId !== item.workspace_id ? workspaceId : undefined,
          name: name.trim(),
          alias_name: aliasName.trim() || undefined,
          match_path: matchPaths,
          description: description.trim() || undefined,
        },
      },
      { onSuccess: () => onOpenChange(false) },
    )
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!updateMutation.isPending) onOpenChange(nextOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>编辑项目</DialogTitle>
          <DialogDescription>
            修改项目信息，或将项目迁移到其他空间。
          </DialogDescription>
        </DialogHeader>
        <fieldset
          disabled={updateMutation.isPending}
          className="grid min-w-0 gap-4 py-4"
        >
          <div className="grid gap-2">
            <Label htmlFor="e-workspace">所属空间 *</Label>
            <Select
              value={workspaceId}
              onValueChange={setWorkspaceId}
              disabled={
                workspaceOptions.isPending ||
                workspaceOptions.isError ||
                updateMutation.isPending
              }
            >
              <SelectTrigger
                id="e-workspace"
                className="w-full min-w-0"
                aria-describedby="e-workspace-help"
              >
                <SelectValue
                  placeholder={
                    workspaceOptions.isPending ? '空间加载中…' : '选择所属空间'
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {workspaces.map((workspace) => (
                  <SelectItem key={workspace.id} value={workspace.id}>
                    <WorkspaceIcon name={workspace.icon} />
                    <span className="truncate">{workspace.name}</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {workspaceOptions.isError ? (
              <div
                role="alert"
                className="flex items-center gap-2 text-sm text-destructive"
              >
                空间加载失败
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void workspaceOptions.refetch()}
                >
                  重试
                </Button>
              </div>
            ) : !workspaceOptions.isPending && workspaces.length === 0 ? (
              <p role="status" className="text-sm text-muted-foreground">
                暂无可用空间。
              </p>
            ) : null}
            <p id="e-workspace-help" className="text-xs text-muted-foreground">
              迁移后，项目及其问答、预览和收到的 Pin 将显示在目标空间，Wiki
              配置保持不变。
            </p>
          </div>
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
        </fieldset>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={updateMutation.isPending}
          >
            取消
          </Button>
          <Button onClick={handleSubmit} disabled={!canSubmit}>
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
