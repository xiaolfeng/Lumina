import { useEffect, useState } from 'react'
import { FolderCode, Plus, X } from 'lucide-react'
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
  const [paths, setPaths] = useState<string[]>([])
  const [newPathInput, setNewPathInput] = useState('')
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
      setPaths(item.match_path ? [...item.match_path] : [])
      setNewPathInput('')
      setDescription(item.description)
      setWorkspaceId(item.workspace_id)
    }
  }, [item, open])

  const handleAddPath = () => {
    const trimmed = newPathInput.trim()
    if (!trimmed) return
    if (!paths.includes(trimmed)) {
      setPaths([...paths, trimmed])
    }
    setNewPathInput('')
  }

  const handleRemovePath = (index: number) => {
    setPaths(paths.filter((_, i) => i !== index))
  }

  const handleSubmit = () => {
    if (!item || !canSubmit) return
    updateMutation.mutate(
      {
        id: item.id,
        data: {
          workspace_id:
            workspaceId !== item.workspace_id ? workspaceId : undefined,
          name: name.trim(),
          alias_name: aliasName.trim() || undefined,
          match_path: paths,
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
          {/* 本地 Agent 检出路径映射 Tag 列表 */}
          <div className="border border-line bg-sand/40 p-3 space-y-2.5">
            <div className="flex items-center gap-1.5 text-xs font-medium text-sea-ink">
              <FolderCode className="size-3.5 text-lagoon" />
              <span>本地工作目录映射 (MatchPath)</span>
            </div>
            <p className="text-[11px] text-sea-ink-soft leading-relaxed">
              供本地 Agent (如 MCP / CLI) 在此目录下工作时自动关联该项目。
            </p>
            <div className="flex gap-2">
              <Input
                value={newPathInput}
                onChange={(e) => setNewPathInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    handleAddPath()
                  }
                }}
                placeholder="如: /Users/username/workspace/repo"
                className="h-8 text-xs font-mono bg-foam"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddPath}
                className="h-8 shrink-0 text-xs px-2.5"
              >
                <Plus className="size-3 mr-1" />
                添加
              </Button>
            </div>
            {paths.length > 0 ? (
              <div className="space-y-1.5 max-h-32 overflow-y-auto">
                {paths.map((p, idx) => (
                  <div
                    key={p}
                    className="flex items-center justify-between gap-2 bg-foam border border-line px-2.5 py-1 text-[11px] font-mono"
                  >
                    <span className="truncate text-sea-ink select-all">
                      {p}
                    </span>
                    <button
                      type="button"
                      onClick={() => handleRemovePath(idx)}
                      className="text-sea-ink-soft hover:text-destructive shrink-0"
                      title="移除此路径"
                    >
                      <X className="size-3" />
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-[11px] text-sea-ink-soft italic">
                暂未绑定路径映射（建议通过 Agent 自动绑定）
              </div>
            )}
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
