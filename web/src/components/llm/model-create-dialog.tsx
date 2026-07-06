import { useState } from 'react'
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
import { useCreateModel, useProviders } from '#/hooks/useLlmConfig'
import { toast } from 'sonner'

interface ModelCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelCreateDialog({
  open,
  onOpenChange,
}: ModelCreateDialogProps) {
  const [providerId, setProviderId] = useState('')
  const [modelName, setModelName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [maxTokens, setMaxTokens] = useState('4096')
  const [temperature, setTemperature] = useState('0.3')
  const [description, setDescription] = useState('')

  const { data: providersData } = useProviders()
  const createMutation = useCreateModel()

  const providers = providersData?.data?.list || []

  const handleSubmit = () => {
    if (!providerId || !modelName.trim() || !displayName.trim()) return
    createMutation.mutate(
      {
        provider_id: providerId,
        model_name: modelName.trim(),
        display_name: displayName.trim(),
        max_tokens: parseInt(maxTokens, 10) || 4096,
        temperature: parseFloat(temperature) || 0.3,
        description: description.trim() || '',
      },
      {
        onSuccess: () => {
          toast.success('Model 创建成功')
          handleClose()
        },
        onError: (error: Error) => {
          toast.error(error.message || '创建失败')
        },
      },
    )
  }

  const handleClose = () => {
    setProviderId('')
    setModelName('')
    setDisplayName('')
    setMaxTokens('4096')
    setTemperature('0.3')
    setDescription('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>创建 Model</DialogTitle>
          <DialogDescription>
            配置一个新的 LLM Model。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="model-provider">Provider *</Label>
            <Select value={providerId} onValueChange={setProviderId}>
              <SelectTrigger id="model-provider">
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
            <Label htmlFor="model-name">模型标识 *</Label>
            <Input
              id="model-name"
              value={modelName}
              onChange={(e) => setModelName(e.target.value)}
              placeholder="如 gpt-4o"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="model-display-name">显示名称 *</Label>
            <Input
              id="model-display-name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="如 GPT-4o"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="model-max-tokens">Max Tokens</Label>
            <Input
              id="model-max-tokens"
              type="number"
              value={maxTokens}
              onChange={(e) => setMaxTokens(e.target.value)}
              placeholder="4096"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="model-temperature">温度</Label>
            <Input
              id="model-temperature"
              type="number"
              step="0.1"
              value={temperature}
              onChange={(e) => setTemperature(e.target.value)}
              placeholder="0.3"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="model-desc">描述</Label>
            <Input
              id="model-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="输入描述（可选）"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              !providerId ||
              !modelName.trim() ||
              !displayName.trim() ||
              createMutation.isPending
            }
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
