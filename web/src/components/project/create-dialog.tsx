import { useState } from 'react'
import { ChevronDown, ChevronRight, FolderCode, Plus, X } from 'lucide-react'
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
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateDialog({ open, onOpenChange }: CreateDialogProps) {
  const [name, setName] = useState('')
  const [aliasName, setAliasName] = useState('')
  const [description, setDescription] = useState('')
  const [paths, setPaths] = useState<string[]>([])
  const [newPathInput, setNewPathInput] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const { current } = useCurrentWorkspace()

  const createMutation = useCreateProject()

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
    if (!name.trim() || !current?.id) return
    createMutation.mutate(
      {
        name: name.trim(),
        alias_name: aliasName.trim() || undefined,
        match_path: paths.length > 0 ? paths : undefined,
        description: description.trim() || undefined,
        workspace_id: current.id,
      },
      {
        onSuccess: () => handleClose(),
      },
    )
  }

  const handleClose = () => {
    setName('')
    setAliasName('')
    setDescription('')
    setPaths([])
    setNewPathInput('')
    setShowAdvanced(false)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <span className="size-2 rotate-45 bg-lagoon" />
            <DialogTitle>新建项目</DialogTitle>
          </div>
          <DialogDescription>
            为当前空间配置新的代码项目，用于组织 RepoWiki、Pin 依赖与 Q&A 会话。
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-2">
          {/* 所属空间提示 */}
          <div className="flex items-center justify-between bg-sand border border-line px-3 py-2 text-xs">
            <span className="text-sea-ink-soft">归属工作空间</span>
            <span className="font-semibold text-sea-ink">
              {current?.name || '默认空间'}
              {current?.slug ? ` (${current.slug})` : ''}
            </span>
          </div>

          {/* 基础表单：项目全称与别名 */}
          <div className="grid gap-2">
            <Label
              htmlFor="p-name"
              className="text-xs font-semibold text-sea-ink"
            >
              项目全名 (Name) <span className="text-destructive">*</span>
            </Label>
            <Input
              id="p-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="如 Lumina / studio"
              className="bg-foam"
            />
          </div>

          <div className="grid gap-2">
            <Label
              htmlFor="p-alias"
              className="text-xs font-semibold text-sea-ink"
            >
              显示别名 (Alias)
            </Label>
            <Input
              id="p-alias"
              value={aliasName}
              onChange={(e) => setAliasName(e.target.value)}
              placeholder="如 微明 / 零号台（用于在台账中优雅呈现）"
              className="bg-foam"
            />
          </div>

          <div className="grid gap-2">
            <Label
              htmlFor="p-desc"
              className="text-xs font-semibold text-sea-ink"
            >
              项目描述
            </Label>
            <textarea
              id="p-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="简要描述项目的职责与业务范畴（可选）"
              className="flex min-h-[72px] w-full rounded-none border border-input bg-foam px-3 py-2 text-xs ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>

          {/* 高级选项：本地 Agent 检出路径映射 */}
          <div className="border border-line bg-sand/40 p-3">
            <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="flex w-full items-center justify-between text-xs font-medium text-sea-ink hover:text-lagoon transition-colors"
            >
              <div className="flex items-center gap-1.5">
                <FolderCode className="size-3.5 text-lagoon" />
                <span>高级设置 · 本地工作目录关联 (MatchPath)</span>
              </div>
              {showAdvanced ? (
                <ChevronDown className="size-3.5 text-sea-ink-soft" />
              ) : (
                <ChevronRight className="size-3.5 text-sea-ink-soft" />
              )}
            </button>

            {showAdvanced ? (
              <div className="mt-3 space-y-2.5 pt-2 border-t border-dashed border-line">
                <p className="text-[11px] text-sea-ink-soft leading-relaxed">
                  供本地 Agent (如 MCP / CLI)
                  在此目录下工作时自动寻址该项目。可留空，后续由 Agent
                  运行时自动关联。
                </p>

                {/* 路径添加输入框 */}
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

                {/* 已添加路径 Tag 胶囊列表 */}
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
                    尚未添加路径映射（建议通过 Agent 自动绑定）
                  </div>
                )}
              </div>
            ) : null}
          </div>
        </div>

        <DialogFooter className="border-t border-line pt-3 mt-1">
          <Button variant="outline" size="sm" onClick={handleClose}>
            取消
          </Button>
          <Button
            size="sm"
            onClick={handleSubmit}
            className="bg-primary text-primary-foreground hover:bg-primary/90"
            disabled={!name.trim() || !current?.id || createMutation.isPending}
          >
            {createMutation.isPending ? '创建中...' : '立即创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
