import { useEffect, useMemo, useState } from 'react'
import { Plus, Search, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Badge } from '@lumina/components/ui/badge'
import { Switch } from '@lumina/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@lumina/components/ui/select'
import { ProviderOption } from '#/components/llm/provider-option'
import {
  useCreateModel,
  useDeleteModel,
  useModels,
  useProviders,
  useUpdateModel,
} from '#/hooks/useLlmConfig'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import type { Model } from '#/lib/models/response/llm'
import { toast } from 'sonner'

export function ModelWorkbench() {
  const { data: modelsData, isLoading: modelsLoading } = useModels()
  const { data: providersData, isLoading: providersLoading } = useProviders()
  const createMutation = useCreateModel()
  const updateMutation = useUpdateModel()
  const deleteMutation = useDeleteModel()

  const rawModels = modelsData?.data?.items ?? []
  const providers = providersData?.data?.items ?? []

  // 当前选中的模型，null 表示「新建模式」
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [modelToDelete, setModelToDelete] = useState<Model | null>(null)

  // 表单状态
  const [providerId, setProviderId] = useState('')
  const [modelName, setModelName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [maxTokens, setMaxTokens] = useState('32000')
  const [contextWindow, setContextWindow] = useState('128000')
  const [temperature, setTemperature] = useState('0.3')
  const [description, setDescription] = useState('')
  const [isActive, setIsActive] = useState(true)

  // 过滤后的模型列表
  const filteredModels = useMemo(() => {
    if (!searchTerm.trim()) return rawModels
    const t = searchTerm.toLowerCase().trim()
    return rawModels.filter(
      (m) =>
        m.model_name.toLowerCase().includes(t) ||
        m.display_name.toLowerCase().includes(t),
    )
  }, [rawModels, searchTerm])

  const selectedModel = useMemo(
    () => rawModels.find((m) => m.id === selectedId) ?? null,
    [rawModels, selectedId],
  )

  // 初始化或切换选中项时回填表单
  useEffect(() => {
    if (selectedModel) {
      setProviderId(selectedModel.provider_id)
      setModelName(selectedModel.model_name)
      setDisplayName(selectedModel.display_name)
      setMaxTokens(String(selectedModel.max_tokens))
      setContextWindow(String(selectedModel.context_window))
      setTemperature(String(selectedModel.temperature))
      setDescription(selectedModel.description || '')
      setIsActive(selectedModel.is_active)
    } else {
      // 默认切换到新建模式，自动绑定第一个 Provider
      setProviderId(providers[0]?.id || '')
      setModelName('')
      setDisplayName('')
      setMaxTokens('32000')
      setContextWindow('128000')
      setTemperature('0.3')
      setDescription('')
      setIsActive(true)
    }
  }, [selectedModel, providers])

  // 默认选中第一个已有模型，若无模型则为新建模式
  useEffect(() => {
    if (rawModels.length > 0 && selectedId === null && !searchTerm) {
      setSelectedId(rawModels[0].id)
    }
  }, [rawModels, selectedId, searchTerm])

  const isCreating = selectedId === null

  const handleStartCreate = () => {
    setSelectedId(null)
    setProviderId(providers[0]?.id || '')
    setModelName('')
    setDisplayName('')
    setMaxTokens('32000')
    setContextWindow('128000')
    setTemperature('0.3')
    setDescription('')
    setIsActive(true)
  }

  const handleSave = () => {
    if (!providerId || !modelName.trim() || !displayName.trim()) {
      toast.error('表单校验未通过', {
        description: 'Provider、模型标识与显示名称为必填项',
      })
      return
    }

    if (isCreating) {
      createMutation.mutate(
        {
          provider_id: providerId,
          model_name: modelName.trim(),
          display_name: displayName.trim(),
          max_tokens: parseInt(maxTokens, 10) || 32000,
          context_window: parseInt(contextWindow, 10) || 128000,
          temperature: parseFloat(temperature) || 0.3,
          description: description.trim() || '',
        },
        {
          onSuccess: (res) => {
            toast.success('模型创建成功', {
              description: `已成功配置模型 [${displayName}]`,
            })
            if (res.data?.id) {
              setSelectedId(res.data.id)
            }
          },
          onError: (err) => {
            toast.error('创建失败', { description: err.message })
          },
        },
      )
    } else if (selectedModel) {
      updateMutation.mutate(
        {
          id: selectedModel.id,
          data: {
            provider_id: providerId,
            model_name: modelName.trim(),
            display_name: displayName.trim(),
            max_tokens: parseInt(maxTokens, 10) || 32000,
            context_window: parseInt(contextWindow, 10) || 128000,
            temperature: parseFloat(temperature) || 0.3,
            is_active: isActive,
            description: description.trim(),
          },
        },
        {
          onSuccess: () => {
            toast.success('模型已更新', { description: `配置已生效同步` })
          },
          onError: (err) => {
            toast.error('更新失败', { description: err.message })
          },
        },
      )
    }
  }

  const handleDeleteClick = (e: React.MouseEvent, m: Model) => {
    e.stopPropagation()
    setModelToDelete(m)
    setDeleteDialogOpen(true)
  }

  if (modelsLoading || providersLoading) {
    return <SkeletonTable />
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
      {/* ── 左侧：模型 Master 列表工作区 (占 5 列) ── */}
      <div className="lg:col-span-5 flex flex-col gap-3">
        {/* 顶部工具栏 */}
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-2.5 size-3.5 text-sea-ink-soft" />
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="搜索模型名称或标识..."
              className="h-8 pl-8 text-xs bg-foam"
            />
          </div>
          <Button
            size="sm"
            onClick={handleStartCreate}
            className={`h-8 shrink-0 text-xs gap-1.5 ${
              isCreating
                ? 'bg-lagoon text-foam'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            }`}
          >
            <Plus className="size-3.5" />
            新建模型
          </Button>
        </div>

        {/* 模型列表项 */}
        <div className="space-y-2 max-h-[calc(100vh-280px)] overflow-y-auto pr-1">
          {filteredModels.length === 0 ? (
            <div className="border border-dashed border-line p-8 text-center bg-foam text-xs text-sea-ink-soft">
              暂无匹配的模型。点击右上角「新建模型」添加。
            </div>
          ) : (
            filteredModels.map((m) => {
              const isSelected = selectedId === m.id
              const p = providers.find((item) => item.id === m.provider_id)
              return (
                <div
                  key={m.id}
                  onClick={() => setSelectedId(m.id)}
                  className={`group relative flex flex-col gap-1.5 p-3.5 border transition-all cursor-pointer ${
                    isSelected
                      ? 'border-lagoon bg-foam shadow-sm border-l-4'
                      : 'border-line bg-surface hover:bg-sand/60'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex flex-col min-w-0">
                      <span className="text-xs font-semibold text-sea-ink truncate">
                        {m.display_name}
                      </span>
                      <span className="font-mono text-[11px] text-sea-ink-soft truncate">
                        {m.model_name}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <Badge
                        variant={m.is_active ? 'default' : 'secondary'}
                        className="text-[10px] px-1.5 py-0"
                      >
                        {m.is_active ? '已启用' : '已停用'}
                      </Badge>
                      <button
                        type="button"
                        onClick={(e) => handleDeleteClick(e, m)}
                        className="opacity-0 group-hover:opacity-100 text-sea-ink-soft hover:text-destructive transition-opacity p-0.5"
                        title="删除模型"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    </div>
                  </div>

                  {/* 胶囊参数行 */}
                  <div className="flex items-center gap-2 pt-1 border-t border-dashed border-line/60 text-[10px] text-sea-ink-soft">
                    <span className="truncate max-w-[120px] font-medium text-lagoon-deep">
                      {p?.name || '未知 Provider'}
                    </span>
                    <span>•</span>
                    <span>{m.context_window.toLocaleString()} ctx</span>
                    <span>•</span>
                    <span>T={m.temperature}</span>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </div>

      {/* ── 右侧：专属编辑与创建工作台 (占 7 列) ── */}
      <div className="lg:col-span-7 border border-line bg-foam p-5 shadow-sm">
        {/* 头部标题区 */}
        <div className="flex items-center justify-between pb-4 mb-5 border-b border-line">
          <div className="flex items-center gap-2">
            <span className="size-2 rotate-45 bg-lagoon" />
            <h3 className="text-sm font-semibold text-sea-ink">
              {isCreating
                ? '新建 LLM 模型配置'
                : `配置模型 · ${selectedModel?.display_name || ''}`}
            </h3>
          </div>
          <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
            {isCreating ? 'CREATE' : 'EDIT'}
          </span>
        </div>

        {/* 表单主体 */}
        <div className="space-y-4">
          <div className="grid gap-1.5">
            <Label className="text-xs font-semibold text-sea-ink">
              归属 Provider <span className="text-destructive">*</span>
            </Label>
            <Select value={providerId} onValueChange={setProviderId}>
              <SelectTrigger className="w-full bg-sand/30">
                <SelectValue placeholder="选择所属供应方" />
              </SelectTrigger>
              <SelectContent>
                {providers.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    <ProviderOption provider={p} />
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                模型标识 (model_name){' '}
                <span className="text-destructive">*</span>
              </Label>
              <Input
                value={modelName}
                onChange={(e) => setModelName(e.target.value)}
                placeholder="如 gpt-4o / claude-3-5-sonnet"
                className="bg-sand/30 font-mono text-xs"
              />
            </div>
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                显示名称 (display_name){' '}
                <span className="text-destructive">*</span>
              </Label>
              <Input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="如 GPT-4o Omni / Claude 3.5 Sonnet"
                className="bg-sand/30 text-xs"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                最大输出 Tokens
              </Label>
              <Input
                type="number"
                value={maxTokens}
                onChange={(e) => setMaxTokens(e.target.value)}
                className="bg-sand/30 font-mono text-xs"
              />
            </div>
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                上下文窗口 (Context)
              </Label>
              <Input
                type="number"
                value={contextWindow}
                onChange={(e) => setContextWindow(e.target.value)}
                className="bg-sand/30 font-mono text-xs"
              />
            </div>
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                温度 (Temperature)
              </Label>
              <Input
                type="number"
                step="0.1"
                min="0"
                max="2"
                value={temperature}
                onChange={(e) => setTemperature(e.target.value)}
                className="bg-sand/30 font-mono text-xs"
              />
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs font-semibold text-sea-ink">
              模型描述 (可选)
            </Label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="输入模型的适用场景或配置说明"
              className="flex min-h-[64px] w-full border border-input bg-sand/30 px-3 py-2 text-xs ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>

          {!isCreating ? (
            <div className="flex items-center justify-between border-t border-line pt-3">
              <div className="flex flex-col">
                <span className="text-xs font-semibold text-sea-ink">
                  模型激活状态
                </span>
                <span className="text-[11px] text-sea-ink-soft">
                  停用后，Agent 将无法调度该模型
                </span>
              </div>
              <Switch checked={isActive} onCheckedChange={setIsActive} />
            </div>
          ) : null}

          {/* 底部专属操作栏 */}
          <div className="flex items-center justify-between border-t border-line pt-4 mt-6">
            <div className="text-[11px] text-sea-ink-soft">
              {isCreating
                ? '新模型将在保存后自动绑定到当前系统目录'
                : `正在编辑 ID: ${selectedModel?.id || ''}`}
            </div>
            <div className="flex items-center gap-2">
              {!isCreating ? (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleStartCreate}
                  className="text-xs"
                >
                  放弃编辑 / 新建
                </Button>
              ) : null}
              <Button
                size="sm"
                onClick={handleSave}
                disabled={
                  !providerId ||
                  !modelName.trim() ||
                  !displayName.trim() ||
                  createMutation.isPending ||
                  updateMutation.isPending
                }
                className="bg-primary text-primary-foreground hover:bg-primary/90 text-xs px-4"
              >
                {createMutation.isPending || updateMutation.isPending
                  ? '保存中...'
                  : isCreating
                    ? '立即创建模型'
                    : '保存模型配置'}
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* 删除确认对话框（仅限破坏性删除） */}
      <ConfirmDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="删除模型"
        description={`确定要删除模型「${modelToDelete?.display_name ?? ''}」吗？此操作不可撤销。`}
        onConfirm={() => {
          if (!modelToDelete) return
          deleteMutation.mutate(modelToDelete.id, {
            onSuccess: () => {
              setDeleteDialogOpen(false)
              if (selectedId === modelToDelete.id) {
                setSelectedId(null)
              }
            },
          })
        }}
        isPending={deleteMutation.isPending}
      />
    </div>
  )
}
