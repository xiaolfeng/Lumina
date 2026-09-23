import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { BookOpen, Check, Copy, ExternalLink, Pencil } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { formatDateTime } from '#/lib/format-date'
import type { ProjectItem } from '#/lib/models/response/project'

interface DetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ProjectItem | null
  wikiStatus?: string
  pendingPinCount?: number
  previewStatus?: string
  onEdit?: (item: ProjectItem) => void
}

export function ProjectDetailSheet({
  open,
  onOpenChange,
  item,
  wikiStatus = '未初始化',
  pendingPinCount = 0,
  previewStatus = '无活跃沙盒',
  onEdit,
}: DetailSheetProps) {
  const navigate = useNavigate()
  const [copiedId, setCopiedId] = useState(false)

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

  const handleEdit = () => {
    onOpenChange(false)
    if (onEdit) onEdit(item)
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
          <div className="flex items-center gap-2">
            <span className="size-2 rotate-45 bg-lagoon" />
            <SheetTitle className="text-base font-semibold text-sea-ink">
              项目完整概览
            </SheetTitle>
          </div>
          <SheetDescription className="text-xs text-sea-ink-soft">
            查看项目的标识别名、底层绝对路径映射与资产配置属性
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5 text-[13px]">
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
              底层匹配路径 (MatchPath)
            </span>
            <p className="text-[11.5px] text-sea-ink-soft">
              Agent 与 MCP 将通过以下本地绝对路径或工作区相对路径自动寻址该项目：
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
                未配置底层路径映射
              </div>
            )}
          </div>

          {/* 关联资产概览 (真实动态传入) */}
          <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
            <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
              核心关联资产状态
            </span>
            <div className="grid grid-cols-3 gap-2 mt-1">
              <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                <span className="text-[10px] uppercase font-bold text-sea-ink-soft">RepoWiki</span>
                <span className="font-mono text-xs font-semibold text-sea-ink truncate" title={wikiStatus}>
                  {wikiStatus}
                </span>
              </div>
              <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                <span className="text-[10px] uppercase font-bold text-sea-ink-soft">Pin 待办</span>
                <span className="font-mono text-xs font-semibold text-sea-ink">
                  {pendingPinCount} 待办
                </span>
              </div>
              <div className="bg-sand border border-line p-2.5 flex flex-col gap-1">
                <span className="text-[10px] uppercase font-bold text-sea-ink-soft">Preview</span>
                <span className="font-mono text-xs font-semibold text-sea-ink truncate" title={previewStatus}>
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
        </div>

        {/* 底部操作工具栏 */}
        <div className="border-t border-line bg-surface-strong px-6 py-4 flex items-center justify-between gap-3">
          <Button
            type="button"
            variant="outline"
            className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
            onClick={handleEdit}
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
        </div>
      </SheetContent>
    </Sheet>
  )
}
