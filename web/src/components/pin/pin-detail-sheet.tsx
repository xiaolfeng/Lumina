import { useEffect, useState } from 'react'
import { Check, Copy, Pencil } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { Label } from '@lumina/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@lumina/components/ui/select'
import { useUpdatePin } from '#/hooks/usePin'
import { useProjectNameMap } from '#/hooks/useProject'
import { formatDateTime } from '#/lib/format-date'
import type { PinItem } from '#/lib/models/response/pin'
import { toast } from 'sonner'

interface PinDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: PinItem | null
  initialMode?: 'view' | 'edit'
}

const CATEGORY_OPTIONS = [
  { value: 'notice', label: '注意事项' },
  { value: 'dependency', label: '依赖约束' },
  { value: 'api_change', label: '接口变更' },
  { value: 'other', label: '其他' },
] as const

const PRIORITY_OPTIONS = [
  { value: 'high', label: '高优先级 (High)' },
  { value: 'medium', label: '中优先级 (Medium)' },
  { value: 'low', label: '低优先级 (Low)' },
] as const

const categoryLabels: Record<string, string> = {
  notice: '注意事项',
  dependency: '依赖约束',
  api_change: '接口变更',
  other: '其他',
}

const priorityLabels: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

export function PinDetailSheet({
  open,
  onOpenChange,
  item,
  initialMode = 'view',
}: PinDetailSheetProps) {
  const [copiedId, setCopiedId] = useState(false)
  const [currentPin, setCurrentPin] = useState<PinItem | null>(item)
  const [isEditing, setIsEditing] = useState(initialMode === 'edit')

  const [category, setCategory] = useState('notice')
  const [priority, setPriority] = useState('medium')

  const { names } = useProjectNameMap()
  const updateMutation = useUpdatePin()

  useEffect(() => {
    if (open && item) {
      setCurrentPin(item)
      setIsEditing(initialMode === 'edit')
      setCategory(item.category || 'notice')
      setPriority(item.priority || 'medium')
      setCopiedId(false)
    }
  }, [open, item, initialMode])

  if (!currentPin) return null

  const fromName =
    names[currentPin.from_project_id] ?? currentPin.from_project_id
  const toName = names[currentPin.to_project_id] ?? currentPin.to_project_id
  const isPending = currentPin.status === 'pending'

  const handleCopyId = () => {
    void navigator.clipboard.writeText(currentPin.id)
    setCopiedId(true)
    setTimeout(() => setCopiedId(false), 2000)
    toast.success('Pin 实体雪花 ID 已复制到剪贴板')
  }

  const handleSaveEdit = () => {
    updateMutation.mutate(
      {
        id: currentPin.id,
        data: {
          category,
          priority,
        },
      },
      {
        onSuccess: () => {
          toast.success('Pin 约束属性已更新')
          setIsEditing(false)
          setCurrentPin((prev) =>
            prev ? { ...prev, category, priority } : null,
          )
        },
        onError: (err) => {
          toast.error('更新失败', { description: err.message })
        },
      },
    )
  }

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
                {isEditing ? '调整约束属性' : 'Pin 约束全景详情'}
              </SheetTitle>
            </div>
            <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
              {isEditing ? 'EDITING' : 'VIEW'}
            </span>
          </div>
          <SheetDescription className="text-xs text-sea-ink-soft">
            {isEditing
              ? '就地修改该约束的分类属性与优先级评级'
              : '查看跨项目依赖约束契约、投递流向与不可变约束正文'}
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5 text-[13px]">
          {isEditing ? (
            /* ── 编辑模式 ── */
            <div className="space-y-4">
              <div className="bg-sand border border-line p-3 text-xs text-sea-ink-soft">
                <span className="font-semibold text-sea-ink">
                  约束不可变原则：
                </span>
                约束标题与正文作为跨项目交付契约已归档上链，仅支持就地调整分发属性与优先级。
              </div>

              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  约束分类 (Category){' '}
                  <span className="text-destructive">*</span>
                </Label>
                <Select value={category} onValueChange={setCategory}>
                  <SelectTrigger className="w-full bg-sand/40 text-xs">
                    <SelectValue placeholder="选择约束分类" />
                  </SelectTrigger>
                  <SelectContent>
                    {CATEGORY_OPTIONS.map((opt) => (
                      <SelectItem
                        key={opt.value}
                        value={opt.value}
                        className="text-xs"
                      >
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-1.5">
                <Label className="text-xs font-semibold text-sea-ink">
                  处理优先级 (Priority){' '}
                  <span className="text-destructive">*</span>
                </Label>
                <Select value={priority} onValueChange={setPriority}>
                  <SelectTrigger className="w-full bg-sand/40 text-xs">
                    <SelectValue placeholder="选择优先级" />
                  </SelectTrigger>
                  <SelectContent>
                    {PRIORITY_OPTIONS.map((opt) => (
                      <SelectItem
                        key={opt.value}
                        value={opt.value}
                        className="text-xs"
                      >
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          ) : (
            /* ── 查看模式 ── */
            <>
              {/* 约束标题与状态 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <div className="flex items-center justify-between">
                  <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                    约束主题 (Title)
                  </span>
                  <span
                    className={`font-mono text-[10.5px] px-2 py-0.5 border ${
                      isPending
                        ? 'border-chip-line bg-lagoon/10 text-lagoon-deep'
                        : 'border-line bg-chip-bg text-sea-ink-soft'
                    }`}
                  >
                    {isPending
                      ? '待下游消费 (PENDING)'
                      : '已消费完成 (CONSUMED)'}
                  </span>
                </div>
                <span className="text-sm font-semibold text-sea-ink">
                  {currentPin.title}
                </span>
              </div>

              {/* 实体雪花 ID */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  实体雪花 ID
                </span>
                <div className="flex items-center justify-between bg-sand border border-line p-2">
                  <span className="font-mono text-xs text-sea-ink select-all">
                    {currentPin.id}
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

              {/* 流向关系 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  依赖流向关系 (Stream)
                </span>
                <div className="flex items-center justify-between bg-sand border border-line p-3 text-xs">
                  <div className="flex flex-col gap-0.5">
                    <span className="text-[10px] text-sea-ink-soft uppercase font-bold">
                      来源项目 (From)
                    </span>
                    <span className="font-semibold text-sea-ink">
                      {fromName}
                    </span>
                  </div>
                  <span className="text-lagoon font-bold text-base">➔</span>
                  <div className="flex flex-col gap-0.5 text-right">
                    <span className="text-[10px] text-sea-ink-soft uppercase font-bold">
                      目标项目 (To)
                    </span>
                    <span className="font-semibold text-sea-ink">{toName}</span>
                  </div>
                </div>
              </div>

              {/* 属性胶囊 */}
              <div className="grid grid-cols-2 gap-3 border-b border-dashed border-line pb-4">
                <div className="bg-sand border border-line p-2.5">
                  <span className="text-[10px] text-sea-ink-soft uppercase font-bold block mb-1">
                    分类
                  </span>
                  <span className="font-semibold text-xs text-sea-ink">
                    {categoryLabels[currentPin.category] ||
                      currentPin.category ||
                      '其他'}
                  </span>
                </div>
                <div className="bg-sand border border-line p-2.5">
                  <span className="text-[10px] text-sea-ink-soft uppercase font-bold block mb-1">
                    优先级
                  </span>
                  <span className="font-semibold text-xs text-sea-ink">
                    {priorityLabels[currentPin.priority] ||
                      currentPin.priority ||
                      '未定义'}
                  </span>
                </div>
              </div>

              {/* 约束正文完整展示 */}
              <div className="flex flex-col gap-1.5 border-b border-dashed border-line pb-4">
                <span className="text-[11px] font-bold uppercase tracking-wider text-lagoon-deep">
                  约束正文内容 (Content)
                </span>
                <div className="border border-line bg-surface p-3.5 text-xs text-sea-ink leading-relaxed whitespace-pre-wrap font-mono break-all max-h-72 overflow-y-auto">
                  {currentPin.content || '（无详细约束正文）'}
                </div>
              </div>

              {/* 时间信息 */}
              <div className="space-y-2 text-xs text-sea-ink-soft">
                <div className="flex justify-between py-1 border-b border-line">
                  <span>创建时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(currentPin.created_at)}
                  </span>
                </div>
                <div className="flex justify-between py-1 border-b border-line">
                  <span>更新时间</span>
                  <span className="font-mono text-sea-ink">
                    {formatDateTime(currentPin.updated_at)}
                  </span>
                </div>
              </div>
            </>
          )}
        </div>

        {/* 底部操作栏 */}
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
                disabled={updateMutation.isPending}
              >
                {updateMutation.isPending ? '保存中...' : '保存更改'}
              </Button>
            </>
          ) : (
            <Button
              type="button"
              variant="outline"
              className="rounded-none border-line text-sea-ink hover:bg-chip-bg w-full"
              onClick={() => setIsEditing(true)}
            >
              <Pencil className="mr-1.5 size-3.5" />
              调整约束属性
            </Button>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
