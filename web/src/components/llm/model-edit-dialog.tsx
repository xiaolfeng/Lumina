import { useEffect, useState } from 'react'
import { Button } from '#/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Switch } from '#/components/ui/switch'
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
  const [maxTokens, setMaxTokens] = useState('4096')
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
          max_tokens: parseInt(maxTokens, 10) || 4096,
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
            <Label htmlFor="edit-model-max-tokens">Max Tokens</Label>
            <Input
              id="edit-model-max-tokens"
              type="number"
              value={maxTokens}
              onChange={(e) => setMaxTokens(e.target.value)}
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
