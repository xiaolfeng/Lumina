import { useState } from 'react'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { useWebhookConfig, useRegenerateWebhook } from '#/hooks/useWebhook'
import { Copy, Check, AlertTriangle, RefreshCw, Shield } from 'lucide-react'
import { toast } from 'sonner'

interface WebhookConfigProps {
  configId: number
}

export function WebhookConfig({ configId }: WebhookConfigProps) {
  const { data, isLoading } = useWebhookConfig(configId)
  const regenerateMutation = useRegenerateWebhook(configId)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [resultOpen, setResultOpen] = useState(false)
  const [copied, setCopied] = useState<string | null>(null)
  const [selectedProvider, setSelectedProvider] = useState('')

  const config = data?.data

  const handleCopy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(label)
      setTimeout(() => setCopied(null), 2000)
      toast.success(`${label} 已复制到剪贴板`)
    } catch {
      toast.error('复制失败')
    }
  }

  const handleRegenerate = () => {
    regenerateMutation.mutate(undefined, {
      onSuccess: () => {
        setConfirmOpen(false)
        setResultOpen(true)
      },
      onError: (error: Error) => {
        toast.error(error.message || '重新生成失败')
      },
    })
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Webhook 配置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="h-8 animate-pulse rounded bg-muted" />
          <div className="h-8 animate-pulse rounded bg-muted" />
          <div className="h-8 animate-pulse rounded bg-muted" />
        </CardContent>
      </Card>
    )
  }

  if (!config) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Webhook 配置</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">加载 Webhook 配置失败</p>
        </CardContent>
      </Card>
    )
  }

  const providers = Object.entries(config.provider_guide)
  const guideText = selectedProvider ? config.provider_guide[selectedProvider] : ''

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Webhook 配置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Webhook URL */}
          <div className="space-y-2">
            <Label>Webhook URL</Label>
            <div className="flex gap-2">
              <Input value={config.url} readOnly className="font-mono text-sm" />
              <Button
                variant="outline"
                size="icon"
                onClick={() => handleCopy(config.url, 'URL')}
              >
                {copied === 'URL' ? (
                  <Check className="size-4 text-emerald-500" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>
          </div>

          {/* Token */}
          <div className="space-y-2">
            <Label>Token</Label>
            <div className="flex items-center gap-2">
              <code className="rounded bg-muted px-2 py-1 font-mono text-sm">
                {config.token}
              </code>
              <Button
                variant="outline"
                size="icon"
                onClick={() => handleCopy(config.token, 'Token')}
              >
                {copied === 'Token' ? (
                  <Check className="size-4 text-emerald-500" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>
          </div>

          {/* Secret status */}
          <div className="flex items-center gap-2">
            <Shield
              className={`size-4 ${config.has_secret ? 'text-emerald-500' : 'text-muted-foreground'}`}
            />
            <span className="text-sm">
              {config.has_secret ? '已设置 Secret' : '未设置 Secret'}
            </span>
          </div>

          {/* Provider Guide */}
          {providers.length > 0 && (
            <div className="space-y-2">
              <Label>提供商配置指南</Label>
              <Select value={selectedProvider} onValueChange={setSelectedProvider}>
                <SelectTrigger>
                  <SelectValue placeholder="选择 Git 提供商" />
                </SelectTrigger>
                <SelectContent>
                  {providers.map(([key]) => (
                    <SelectItem key={key} value={key}>
                      {key}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {guideText && (
                <div className="rounded-md border bg-muted/50 p-3">
                  <pre className="whitespace-pre-wrap text-sm leading-relaxed">{guideText}</pre>
                </div>
              )}
            </div>
          )}

          {/* Regenerate button */}
          <div className="pt-2">
            <Button
              variant="outline"
              onClick={() => setConfirmOpen(true)}
              className="gap-2"
            >
              <RefreshCw className="size-4" />
              重新生成凭据
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Confirm dialog */}
      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="size-5 text-amber-500" />
              确认重新生成？
            </DialogTitle>
            <DialogDescription>
              重新生成后，旧的 Token 和 Secret 将立即失效。所有使用旧凭据的 Webhook 请求都会被拒绝。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              取消
            </Button>
            <Button
              onClick={handleRegenerate}
              disabled={regenerateMutation.isPending}
              className="gap-2"
            >
              {regenerateMutation.isPending ? (
                <>
                  <RefreshCw className="size-4 animate-spin" />
                  生成中...
                </>
              ) : (
                '确认重新生成'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Result dialog */}
      <Dialog open={resultOpen} onOpenChange={setResultOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>凭据已重新生成</DialogTitle>
            <DialogDescription>
              请立即保存以下信息，关闭后将无法再次查看完整凭据。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Token</Label>
              <div className="flex gap-2">
                <Input
                  value={regenerateMutation.data?.data?.token ?? ''}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() =>
                    handleCopy(regenerateMutation.data?.data?.token ?? '', '新 Token')
                  }
                >
                  {copied === '新 Token' ? (
                    <Check className="size-4 text-emerald-500" />
                  ) : (
                    <Copy className="size-4" />
                  )}
                </Button>
              </div>
            </div>
            <div className="space-y-2">
              <Label>Secret</Label>
              <div className="flex gap-2">
                <Input
                  value={regenerateMutation.data?.data?.secret ?? ''}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() =>
                    handleCopy(regenerateMutation.data?.data?.secret ?? '', '新 Secret')
                  }
                >
                  {copied === '新 Secret' ? (
                    <Check className="size-4 text-emerald-500" />
                  ) : (
                    <Copy className="size-4" />
                  )}
                </Button>
              </div>
            </div>
            <div className="space-y-2">
              <Label>Webhook URL</Label>
              <div className="flex gap-2">
                <Input
                  value={regenerateMutation.data?.data?.url ?? ''}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() =>
                    handleCopy(regenerateMutation.data?.data?.url ?? '', '新 URL')
                  }
                >
                  {copied === '新 URL' ? (
                    <Check className="size-4 text-emerald-500" />
                  ) : (
                    <Copy className="size-4" />
                  )}
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setResultOpen(false)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
