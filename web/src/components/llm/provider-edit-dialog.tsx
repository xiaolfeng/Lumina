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
import { useUpdateProvider } from '#/hooks/useLlmConfig'
import type { Provider } from '#/lib/models/response/llm'
import { toast } from 'sonner'

interface ProviderEditDialogProps {
  item: Provider | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ProviderEditDialog({
  item,
  open,
  onOpenChange,
}: ProviderEditDialogProps) {
  const [name, setName] = useState('')
  const [protocol, setProtocol] = useState('openai')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [description, setDescription] = useState('')
  const [isActive, setIsActive] = useState(true)

  const updateMutation = useUpdateProvider()

  useEffect(() => {
    if (item) {
      setName(item.name)
      setProtocol(item.protocol)
      setBaseUrl(item.base_url || '')
      setApiKey('')
      setDescription(item.description || '')
      setIsActive(item.is_active)
    }
  }, [item])

  const handleSubmit = () => {
    if (!item || !name.trim()) return
    const data: {
      name: string
      protocol: string
      base_url: string
      api_key?: string
      is_active: boolean
      description: string
    } = {
      name: name.trim(),
      protocol,
      base_url: baseUrl.trim(),
      is_active: isActive,
      description: description.trim(),
    }
    if (apiKey.trim()) {
      data.api_key = apiKey.trim()
    }
    updateMutation.mutate(
      { id: item.id, data },
      {
        onSuccess: () => {
          toast.success('Provider 更新成功')
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
          <DialogTitle>编辑 Provider</DialogTitle>
          <DialogDescription>
            修改 Provider 的配置信息。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="edit-provider-name">名称 *</Label>
            <Input
              id="edit-provider-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-provider-protocol">协议 *</Label>
            <Select value={protocol} onValueChange={setProtocol}>
              <SelectTrigger id="edit-provider-protocol">
                <SelectValue placeholder="选择协议" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="openai">OpenAI</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-provider-base-url">Base URL</Label>
            <Input
              id="edit-provider-base-url"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-provider-api-key">
              API Key{' '}
              <span className="text-muted-foreground text-xs">
                （留空不修改）
              </span>
            </Label>
            <Input
              id="edit-provider-api-key"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="输入新 API Key（留空不修改）"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-provider-desc">描述</Label>
            <Input
              id="edit-provider-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="edit-provider-active">启用状态</Label>
            <Switch
              id="edit-provider-active"
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
            disabled={!name.trim() || updateMutation.isPending}
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
