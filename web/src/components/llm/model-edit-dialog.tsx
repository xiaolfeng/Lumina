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
import { Switch } from '@lumina/components/ui/switch'
import { useUpdateModel, useProviders } from '#/hooks/useLlmConfig'
import type { Model } from '#/lib/models/response/llm'
import { toast } from 'sonner'

interface ModelEditDialogProps {
  item: Model | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelEditDialog({
  item,
  open,
  onOpenChange,
}: ModelEditDialogProps) {
  const [providerId, setProviderId] = useState('')
  const [modelName, setModelName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [maxTokens, setMaxTokens] = useState('32000')
  const [contextWindow, setContextWindow] = useState('128000')
  const [temperature, setTemperature] = useState('0.3')
  const [description, setDescription] = useState('')
  const [isActive, setIsActive] = useState(true)

  const { data: providersData } = useProviders()
  const updateMutation = useUpdateModel()

  const providers = providersData?.data?.items ?? []

  useEffect(() => {
    if (item) {
      setProviderId(item.provider_id)
      setModelName(item.model_name)
      setDisplayName(item.display_name)
      setMaxTokens(String(item.max_tokens))
      setContextWindow(String(item.context_window))
      setTemperature(String(item.temperature))
      setDescription(item.description || '')
      setIsActive(item.is_active)
    }
  }, [item])

  const handleSubmit = () => {
    if (!item || !modelName.trim() || !displayName.trim()) return
    updateMutation.mutate(
      {
        id: item.id,
        data: {
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
          toast.success('Model 更新成功')
          onOpenChange(false)
        },
        onError: (error: Error) => {
          toast.error(error.message || '更新失败')
        },
      },
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑 Model</DialogTitle>
          <DialogDescription>
            修改 Model 的配置信息。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="edit-model-provider">Provider</Label>
            <Select value={providerId} onValueChange={setProviderId}>
              <SelectTrigger id="edit-model-provider">
                <SelectValue placeholder="选择 Provider" />
              </SelectTrigger>
              <SelectContent>
                {providers.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-name">模型标识 *</Label>
            <Input
              id="edit-model-name"
              value={modelName}
              onChange={(e) => setModelName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-display-name">显示名称 *</Label>
            <Input
              id="edit-model-display-name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-max-tokens">最大输出 Tokens</Label>
            <Input
              id="edit-model-max-tokens"
              type="number"
              value={maxTokens}
              onChange={(e) => setMaxTokens(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-context-window">上下文窗口</Label>
            <Input
              id="edit-model-context-window"
              type="number"
              value={contextWindow}
              onChange={(e) => setContextWindow(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-temperature">温度</Label>
            <Input
              id="edit-model-temperature"
              type="number"
              step="0.1"
              value={temperature}
              onChange={(e) => setTemperature(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-model-desc">描述</Label>
            <Input
              id="edit-model-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="edit-model-active">启用状态</Label>
            <Switch
              id="edit-model-active"
              checked={isActive}
              onCheckedChange={setIsActive}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              !modelName.trim() ||
              !displayName.trim() ||
              updateMutation.isPending
            }
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
