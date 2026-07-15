import { useState } from 'react'
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
import { useCreateProvider } from '#/hooks/useLlmConfig'
import { toast } from 'sonner'

interface ProviderCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ProviderCreateDialog({
  open,
  onOpenChange,
}: ProviderCreateDialogProps) {
  const [name, setName] = useState('')
  const [protocol, setProtocol] = useState('openai')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [description, setDescription] = useState('')

  const createMutation = useCreateProvider()

  const handleSubmit = () => {
    if (!name.trim() || !apiKey.trim()) return
    createMutation.mutate(
      {
        name: name.trim(),
        protocol,
        base_url: baseUrl.trim() || '',
        api_key: apiKey.trim(),
        description: description.trim() || '',
      },
      {
        onSuccess: () => {
          toast.success('Provider 创建成功')
          handleClose()
        },
        onError: (error: Error) => {
          toast.error(error.message || '创建失败')
        },
      },
    )
  }

  const handleClose = () => {
    setName('')
    setProtocol('openai')
    setBaseUrl('')
    setApiKey('')
    setDescription('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>创建 Provider</DialogTitle>
          <DialogDescription>
            配置一个新的 LLM Provider。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="provider-name">名称 *</Label>
            <Input
              id="provider-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="输入 Provider 名称"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="provider-protocol">协议 *</Label>
            <Select value={protocol} onValueChange={setProtocol}>
              <SelectTrigger id="provider-protocol">
                <SelectValue placeholder="选择协议" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="openai">OpenAI</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="provider-base-url">Base URL</Label>
            <Input
              id="provider-base-url"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="https://api.openai.com/v1（可选）"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="provider-api-key">API Key *</Label>
            <Input
              id="provider-api-key"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="输入 API Key"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="provider-desc">描述</Label>
            <Input
              id="provider-desc"
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
            disabled={!name.trim() || !apiKey.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
