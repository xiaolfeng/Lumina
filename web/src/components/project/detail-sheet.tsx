import { useEffect, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  BookOpen,
  Check,
  Copy,
  ExternalLink,
  FolderCode,
  Pencil,
  Plus,
  X,
} from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
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
import { formatDateTime } from '#/lib/format-date'
import type { ProjectItem } from '#/lib/models/response/project'
import { toast } from 'sonner'

interface DetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ProjectItem | null
  wikiStatus?: string
  pendingPinCount?: number
  previewStatus?: string
  initialMode?: 'view' | 'edit'
}

export function ProjectDetailSheet({
  open,
  onOpenChange,
  item,
  wikiStatus = '未初始化',
  pendingPinCount = 0,
  previewStatus = '无活跃沙盒',
  initialMode = 'view',
}: DetailSheetProps) {
  const navigate = useNavigate()
  const [copiedId, setCopiedId] = useState(false)
  const [isEditing, setIsEditing] = useState(initialMode === 'edit')

  // 编辑表单状态
  const [name, setName] = useState('')
  const [aliasName, setAliasName] = useState('')
  const [description, setDescription] = useState('')
  const [workspaceId, setWorkspaceId] = useState('')
  const [paths, setPaths] = useState<string[]>([])
  const [newPathInput, setNewPathInput] = useState('')

  const workspaceOptions = useWorkspaceOptions()
  const workspaces = workspaceOptions.data ?? []
  const updateMutation = useUpdateProject()

  useEffect(() => {
    if (open && item) {
      setIsEditing(initialMode === 'edit')
      setName(item.name)
      setAliasName(item.alias_name || '')
      setDescription(item.description || '')
      setWorkspaceId(item.workspace_id)
      setPaths(item.match_path ? [...item.match_path] : [])
      setNewPathInput('')
    }
  }, [open, item, initialMode])

  if (!item) return null

  const handleCopyId = () => {
    void navigator.clipboard.writeText(item.id)
    setCopiedId(true)
    setTimeout(() => setCopiedId(false), 2000)
  }

  const handleOpenWiki = () => {
    onOpenChange(false)
    void navigate({
      to: '/console/project/$projectId/repowiki',
      params: { projectId: item.id },
    })
  }

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

  const handleSaveEdit = () => {
    if (!name.trim() || !workspaceId) return
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
      {
        onSuccess: () => {
          toast.success('项目配置已更新')
          setIsEditing(false)
        },
        onError: (err) => {
          toast.error('更新失败', { description: err.message })
        },
      },
    )
  }

  const matchPaths =
    item.match_path && item.match_path.length > 0 ? item.match_path : []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="w-full rounded-none border-l border-line bg-foam p-0 sm:max-w-lg flex flex-col"
      >
        <SheetHeader className="border-b border-line bg-surface-strong px-6 py-5 text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="size-2 rotate-45 bg-lagoon" />
              <SheetTitle className="text-base font-semibold text-sea-ink">
                {isEditing ? '编辑项目配置' : '项目完整概览'}
              </SheetTitle>
            </div>
            <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
              {isEditing ? 'EDITING' : 'VIEW'}
            </span>
          </div>
          <SheetDescription className="text-xs text-sea-ink-soft">
            {isEditing
              ? '修改项目基础资料、所属空间以及本地 Agent 检出路径映射'
              : '查看项目的标识别名、底层路径映射与核心关联资产状态'}
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5 text-[13px]">
          {isEditing ? (
            /* ── 编辑模式专属工作区 ── */
            <div className="space-y-4">
              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  所属空间 <span className="text-destructive">*</span>
                </Label>
                <Select value={workspaceId} onValueChange={setWorkspaceId}>
                  <SelectTrigger className="w-full bg-sand/40">
                    <SelectValue placeholder="选择所属空间" />
                  </SelectTrigger>
                  <SelectContent>
                    {workspaces.map((ws) => (
                      <SelectItem key={ws.id} value={ws.id}>
                        <div className="flex items-center gap-2">
                          <WorkspaceIcon name={ws.icon} />
                          <span>{ws.name}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  项目全名 (Name) <span className="text-destructive">*</span>
                </Label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="项目全称"
                  className="bg-sand/40 text-xs"
                />
              </div>

              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  显示别名 (Alias)
                </Label>
                <Input
                  value={aliasName}
                  onChange={(e) => setAliasName(e.target.value)}
                  placeholder="项目别名（用于台账优雅呈现）"
                  className="bg-sand/40 text-xs"
                />
              </div>

              {/* MatchPath Tag 列表 */}
              <div className="border border-line bg-sand/40 p-3 space-y-2.5">
                <div className="flex items-center gap-1.5 text-xs font-medium text-sea-ink">
                  <FolderCode className="size-3.5 text-lagoon" />
                  <span>本地工作目录映射 (MatchPath)</span>
                </div>
                <p className="text-[11px] text-sea-ink-soft leading-relaxed">
                  供本地 Agent (如 MCP / CLI) 在此目录下工作时自动寻址该项目。
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
                          title="移除"
                        >
                          <X className="size-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-[11px] text-sea-ink-soft italic">
                    暂未配置路径映射
                  </div>
                )}
              </div>

              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  项目介绍
                </Label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="简述该项目职责"
                  className="flex min-h-[72px] w-full border border-input bg-sand/40 px-3 py-2 text-xs ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
            </div>
          ) : (
            /* ── 查看模式 ── */
            <>
              {/* 项目别名 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  项目别名 (Alias)
                </span>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm font-semibold text-sea-ink bg-sand border border-line px-2 py-0.5">
                    {item.alias_name || '未配置'}
                  </span>
                </div>
              </div>

              {/* 实际项目全称 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  实际项目全称 (Name)
                </span>
                <span className="text-sm font-semibold text-sea-ink">
                  {item.name}
                </span>
              </div>

              {/* 雪花 ID */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  实体雪花 ID
                </span>
                <div className="flex items-center justify-between bg-sand border border-line p-2">
                  <span className="font-mono text-xs text-sea-ink select-all">
                    {item.id}
                  </span>
                  <button
                    type="button"
                    onClick={handleCopyId}
                    className="inline-flex items-center gap-1 text-[11px] text-sea-ink-soft hover:text-lagoon transition-colors"
                    title="复制 ID"
                  >
                    {copiedId ? (
                      <>
                        <Check className="size-3 text-lagoon" />
                        <span>已复制</span>
                      </>
                    ) : (
                      <>
                        <Copy className="size-3" />
                        <span>复制</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              {/* 项目描述 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  项目介绍 (Description)
                </span>
                <p className="text-xs text-sea-ink-soft leading-relaxed">
                  {item.description || '暂无详细描述'}
                </p>
              </div>

              {/* 底层路径映射 MatchPath */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  本地工作目录映射 (MatchPath)
                </span>
                <p className="text-[11.5px] text-sea-ink-soft">
                  Agent 与 MCP 通过以下本地绝对路径自动寻址该项目：
                </p>
                {matchPaths.length > 0 ? (
                  <div className="space-y-1.5 mt-1">
                    {matchPaths.map((path) => (
                      <div
                        key={path}
                        className="font-mono text-xs bg-sand border border-line px-2.5 py-1.5 text-sea-ink break-all select-all"
                      >
                        {path}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="font-mono text-xs bg-sand border border-dashed border-line px-2.5 py-2 text-sea-ink-soft">
                    未配置本地路径映射（建议通过 Agent 自动绑定）
                  </div>
                )}
              </div>

              {/* 关联资产概览 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  核心关联资产状态
                </span>
                <div className="grid grid-cols-3 gap-2 mt-1">
                  <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                    <span className="text-[10px] uppercase font-bold text-sea-ink-soft">
                      RepoWiki
                    </span>
                    <span
                      className="font-mono text-xs font-semibold text-sea-ink truncate"
                      title={wikiStatus}
                    >
                      {wikiStatus}
                    </span>
                  </div>
                  <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                    <span className="text-[10px] uppercase font-bold text-sea-ink-soft">
                      Pin 待办
                    </span>
                    <span className="font-mono text-xs font-semibold text-sea-ink">
                      {pendingPinCount} 待办
                    </span>
                  </div>
                  <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                    <span className="text-[10px] uppercase font-bold text-sea-ink-soft">
                      Preview
                    </span>
                    <span
                      className="font-mono text-xs font-semibold text-sea-ink truncate"
                      title={previewStatus}
                    >
                      {previewStatus}
                    </span>
                  </div>
                </div>
              </div>

              {/* 资产与时间信息 */}
              <div className="space-y-2 text-xs text-sea-ink-soft">
                <div className="flex justify-between py-1 border-b border-line">
                  <span>工作空间 ID</span>
                  <span className="font-mono text-sea-ink">
                    {item.workspace_id}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b border-line">
                  <span>创建时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(item.created_at)}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b border-line">
                  <span>更新时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(item.updated_at)}
                  </span>
                </div>
              </div>
            </>
          )}
        </div>

        {/* 底部操作工具栏 */}
        <div className="border-t border-line bg-surface-strong px-6 py-4 flex items-center justify-between gap-3">
          {isEditing ? (
            <>
              <Button
                type="button"
                variant="outline"
                className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
                onClick={() => setIsEditing(false)}
                disabled={updateMutation.isPending}
              >
                取消
              </Button>
              <Button
                type="button"
                className="rounded-none bg-primary text-primary-foreground hover:bg-primary/90 flex-1"
                onClick={handleSaveEdit}
                disabled={
                  !name.trim() || !workspaceId || updateMutation.isPending
                }
              >
                {updateMutation.isPending ? '保存中...' : '保存更改'}
              </Button>
            </>
          ) : (
            <>
              <Button
                type="button"
                variant="outline"
                className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
                onClick={() => setIsEditing(true)}
              >
                <Pencil className="mr-1.5 size-3.5" />
                编辑配置
              </Button>
              <Button
                type="button"
                className="rounded-none bg-lagoon text-foam hover:bg-lagoon-deep flex-1"
                onClick={handleOpenWiki}
              >
                <BookOpen className="mr-1.5 size-3.5" />
                知识库 Wiki
                <ExternalLink className="ml-1 size-3" />
              </Button>
            </>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
