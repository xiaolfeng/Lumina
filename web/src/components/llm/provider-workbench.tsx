import { useEffect, useMemo, useRef, useState } from 'react'
import { Plus, Search, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Badge } from '@lumina/components/ui/badge'
import { Switch } from '@lumina/components/ui/switch'
import {
  useCreateProvider,
  useDeleteProvider,
  useProviders,
  useUpdateProvider,
} from '#/hooks/useLlmConfig'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import type { Provider } from '#/lib/models/response/llm'
import { toast } from 'sonner'

export function ProviderWorkbench() {
  const { data: providersData, isLoading: providersLoading } = useProviders()
  const createMutation = useCreateProvider()
  const updateMutation = useUpdateProvider()
  const deleteMutation = useDeleteProvider()

  const rawProviders = providersData?.data?.items ?? []

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [providerToDelete, setProviderToDelete] = useState<Provider | null>(
    null,
  )

  // 标记是否已执行过初次装载默认选中
  const hasInitialized = useRef(false)

  // 表单状态
  const [name, setName] = useState('')
  const [protocol, setProtocol] = useState('openai')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [description, setDescription] = useState('')
  const [isActive, setIsActive] = useState(true)

  const filteredProviders = useMemo(() => {
    if (!searchTerm.trim()) return rawProviders
    const t = searchTerm.toLowerCase().trim()
    return rawProviders.filter(
      (p) =>
        p.name.toLowerCase().includes(t) ||
        p.protocol.toLowerCase().includes(t),
    )
  }, [rawProviders, searchTerm])

  const selectedProvider = useMemo(
    () => rawProviders.find((p) => p.id === selectedId) ?? null,
    [rawProviders, selectedId],
  )

  useEffect(() => {
    if (!isCreating && selectedProvider) {
      setName(selectedProvider.name)
      setProtocol(selectedProvider.protocol)
      setBaseUrl(selectedProvider.base_url || '')
      setApiKey('') // 密钥不回显，留空表示不修改
      setDescription(selectedProvider.description || '')
      setIsActive(selectedProvider.is_active)
    } else if (isCreating) {
      setName('')
      setProtocol('openai')
      setBaseUrl('https://api.openai.com/v1')
      setApiKey('')
      setDescription('')
      setIsActive(true)
    }
  }, [selectedProvider, isCreating])

  // 仅在首次拉取到数据时默认选中第一个
  useEffect(() => {
    if (!hasInitialized.current && rawProviders.length > 0) {
      setSelectedId(rawProviders[0].id)
      setIsCreating(false)
      hasInitialized.current = true
    } else if (rawProviders.length === 0 && !isCreating) {
      setIsCreating(true)
      setSelectedId(null)
    }
  }, [rawProviders, isCreating])

  const handleStartCreate = () => {
    setIsCreating(true)
    setSelectedId(null)
    setName('')
    setProtocol('openai')
    setBaseUrl('https://api.openai.com/v1')
    setApiKey('')
    setDescription('')
    setIsActive(true)
  }

  const handleSelectProvider = (id: string) => {
    setIsCreating(false)
    setSelectedId(id)
  }

  const handleCancelCreate = () => {
    setIsCreating(false)
    if (rawProviders.length > 0) {
      setSelectedId(rawProviders[0].id)
    }
  }

  const handleSave = () => {
    if (!name.trim() || !protocol.trim()) {
      toast.error('表单校验未通过', { description: '名称与协议为必填项' })
      return
    }

    if (isCreating) {
      if (!apiKey.trim()) {
        toast.error('表单校验未通过', {
          description: '创建 Provider 时 API 密钥为必填项',
        })
        return
      }
      createMutation.mutate(
        {
          name: name.trim(),
          protocol: protocol.trim(),
          base_url: baseUrl.trim(),
          api_key: apiKey.trim(),
          description: description.trim(),
        },
        {
          onSuccess: (res) => {
            toast.success('Provider 创建成功', {
              description: `已成功接入 ${name}`,
            })
            setIsCreating(false)
            if (res.data?.id) {
              setSelectedId(res.data.id)
            }
          },
          onError: (err) => {
            toast.error('创建失败', { description: err.message })
          },
        },
      )
    } else if (selectedProvider) {
      updateMutation.mutate(
        {
          id: selectedProvider.id,
          data: {
            name: name.trim(),
            protocol: protocol.trim(),
            base_url: baseUrl.trim() || undefined,
            api_key: apiKey.trim() || undefined,
            is_active: isActive,
            description: description.trim() || undefined,
          },
        },
        {
          onSuccess: () => {
            toast.success('Provider 已更新', { description: '配置已生效保存' })
          },
          onError: (err) => {
            toast.error('更新失败', { description: err.message })
          },
        },
      )
    }
  }

  const handleDeleteClick = (e: React.MouseEvent, p: Provider) => {
    e.stopPropagation()
    setProviderToDelete(p)
    setDeleteDialogOpen(true)
  }

  if (providersLoading) {
    return <SkeletonTable />
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
      {/* ── 左侧：Provider Master 列表 (占 5 列) ── */}
      <div className="lg:col-span-5 flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-2.5 size-3.5 text-sea-ink-soft" />
            <Input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="搜索供应方名称或协议..."
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
            新建 Provider
          </Button>
        </div>

        <div className="space-y-2 max-h-[calc(100vh-280px)] overflow-y-auto pr-1">
          {filteredProviders.length === 0 ? (
            <div className="border border-dashed border-line p-8 text-center bg-foam text-xs text-sea-ink-soft">
              暂无匹配的 Provider。点击右上角新建。
            </div>
          ) : (
            filteredProviders.map((p) => {
              const isSelected = selectedId === p.id
              return (
                <div
                  key={p.id}
                  onClick={() => handleSelectProvider(p.id)}
                  className={`group relative flex flex-col gap-1.5 p-3.5 border transition-all cursor-pointer ${
                    isSelected
                      ? 'border-lagoon bg-foam shadow-sm border-l-4'
                      : 'border-line bg-surface hover:bg-sand/60'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex flex-col min-w-0">
                      <span className="text-xs font-semibold text-sea-ink truncate">
                        {p.name}
                      </span>
                      <span className="font-mono text-[11px] text-sea-ink-soft truncate">
                        {p.base_url || '默认 BaseURL'}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <Badge
                        variant={p.is_active ? 'default' : 'secondary'}
                        className="text-[10px] px-1.5 py-0"
                      >
                        {p.is_active ? '启用' : '停用'}
                      </Badge>
                      <button
                        type="button"
                        onClick={(e) => handleDeleteClick(e, p)}
                        className="opacity-0 group-hover:opacity-100 text-sea-ink-soft hover:text-destructive transition-opacity p-0.5"
                        title="删除 Provider"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-1 border-t border-dashed border-line/60 text-[10px] text-sea-ink-soft font-mono">
                    <span className="font-semibold text-lagoon-deep uppercase">
                      {p.protocol}
                    </span>
                    <span>•</span>
                    <span>{p.has_key ? '已加密存盘' : '无密钥'}</span>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </div>

      {/* ── 右侧：专属编辑与创建工作台 (占 7 列) ── */}
      <div className="lg:col-span-7 border border-line bg-foam p-5 shadow-sm">
        <div className="flex items-center justify-between pb-4 mb-5 border-b border-line">
          <div className="flex items-center gap-2">
            <span className="size-2 rotate-45 bg-lagoon" />
            <h3 className="text-sm font-semibold text-sea-ink">
              {isCreating
                ? '接入新的 LLM 供应方'
                : `配置供应方 · ${selectedProvider?.name || ''}`}
            </h3>
          </div>
          <div className="flex items-center gap-2">
            {isCreating && rawProviders.length > 0 && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleCancelCreate}
                className="h-7 text-xs rounded-none border-line text-sea-ink hover:bg-chip-bg"
              >
                取消新建
              </Button>
            )}
            <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
              {isCreating ? 'CREATE' : 'EDIT'}
            </span>
          </div>
        </div>

        <div className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                供应方名称 <span className="text-destructive">*</span>
              </Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="如 OpenAI / Anthropic / DeepSeek"
                className="bg-sand/30 text-xs"
              />
            </div>
            <div className="grid gap-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                接口协议 (Protocol) <span className="text-destructive">*</span>
              </Label>
              <Input
                value={protocol}
                onChange={(e) => setProtocol(e.target.value)}
                placeholder="如 openai / anthropic / ollama"
                className="bg-sand/30 font-mono text-xs"
              />
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs font-semibold text-sea-ink">
              Base URL 端点
            </Label>
            <Input
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="如 https://api.openai.com/v1"
              className="bg-sand/30 font-mono text-xs"
            />
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs font-semibold text-sea-ink">
              API Key (密钥){' '}
              {isCreating ? <span className="text-destructive">*</span> : null}
            </Label>
            <Input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder={
                isCreating
                  ? '输入 API Key（将经 AES-256-GCM 加密落库）'
                  : '留空表示保持现有加密密钥不变'
              }
              className="bg-sand/30 font-mono text-xs"
            />
            <p className="text-[11px] text-sea-ink-soft">
              密钥将经由系统 AES-256-GCM 加密存储，绝不以明文暴露。
            </p>
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs font-semibold text-sea-ink">
              描述 (可选)
            </Label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="输入供应方的相关说明"
              className="flex min-h-[64px] w-full border border-input bg-sand/30 px-3 py-2 text-xs ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>

          {!isCreating ? (
            <div className="flex items-center justify-between border-t border-line pt-3">
              <div className="flex flex-col">
                <span className="text-xs font-semibold text-sea-ink">
                  供应方激活状态
                </span>
                <span className="text-[11px] text-sea-ink-soft">
                  停用后，其下所有模型将自动处于不可调度状态
                </span>
              </div>
              <Switch checked={isActive} onCheckedChange={setIsActive} />
            </div>
          ) : null}

          <div className="flex items-center justify-between border-t border-line pt-4 mt-6">
            <div className="text-[11px] text-sea-ink-soft">
              {isCreating
                ? '创建后可立即为其添加专属模型'
                : `ID: ${selectedProvider?.id || ''}`}
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
                  !name.trim() ||
                  !protocol.trim() ||
                  (isCreating && !apiKey.trim()) ||
                  createMutation.isPending ||
                  updateMutation.isPending
                }
                className="bg-primary text-primary-foreground hover:bg-primary/90 text-xs px-4"
              >
                {createMutation.isPending || updateMutation.isPending
                  ? '保存中...'
                  : isCreating
                    ? '立即创建 Provider'
                    : '保存配置'}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <ConfirmDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="删除 Provider"
        description={`确定要删除 Provider「${providerToDelete?.name ?? ''}」吗？此操作不可撤销。`}
        onConfirm={() => {
          if (!providerToDelete) return
          deleteMutation.mutate(providerToDelete.id, {
            onSuccess: () => {
              setDeleteDialogOpen(false)
              if (selectedId === providerToDelete.id) {
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
