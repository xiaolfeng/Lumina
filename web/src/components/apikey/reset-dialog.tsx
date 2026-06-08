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
import { useResetApikey } from '#/hooks/useApikey'
import type {
  ApikeyItem,
  ApikeyResetResponse,
} from '#/lib/models/response/apikey'
import { Copy, Check, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

interface ResetDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApikeyItem | null
}

export function ResetDialog({ open, onOpenChange, item }: ResetDialogProps) {
  const [result, setResult] = useState<ApikeyResetResponse | null>(null)
  const [copied, setCopied] = useState(false)

  const resetMutation = useResetApikey()

  const handleReset = () => {
    if (!item) return
    resetMutation.mutate(item.id, {
      onSuccess: (res) => {
        if (res.data) setResult(res.data)
        toast.success('密钥已重置')
      },
    })
  }

  const handleCopy = async () => {
    if (!result?.key) return
    await navigator.clipboard.writeText(result.key)
    setCopied(true)
    toast.success('已复制到剪贴板')
    setTimeout(() => setCopied(false), 2000)
  }

  const handleClose = () => {
    setResult(null)
    setCopied(false)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        {!result ? (
          <>
            <DialogHeader>
              <DialogTitle>重置密钥</DialogTitle>
              <DialogDescription>
                为 &quot;{item?.name}&quot; 生成新的 API 密钥。
              </DialogDescription>
            </DialogHeader>
            <div className="flex items-start gap-3 rounded-md bg-amber-50 p-4 dark:bg-amber-950/20">
              <AlertTriangle className="mt-0.5 size-5 text-amber-600" />
              <div className="text-sm text-amber-800 dark:text-amber-200">
                <p className="font-medium">警告</p>
                <p>
                  重置后旧密钥将立即失效，所有使用旧密钥的服务将无法访问。
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={handleClose}>
                取消
              </Button>
              <Button
                onClick={handleReset}
                disabled={resetMutation.isPending}
              >
                {resetMutation.isPending ? '重置中...' : '确认重置'}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>密钥已重置</DialogTitle>
              <DialogDescription>
                请立即复制并安全保存此新密钥，关闭后将无法再次查看。
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="rounded-md bg-muted p-4">
                <code className="break-all text-sm font-mono">
                  {result.key}
                </code>
              </div>
              <Button
                variant="outline"
                className="w-full"
                onClick={handleCopy}
              >
                {copied ? (
                  <Check className="mr-2 size-4" />
                ) : (
                  <Copy className="mr-2 size-4" />
                )}
                {copied ? '已复制' : '复制到剪贴板'}
              </Button>
            </div>
            <DialogFooter>
              <Button onClick={handleClose} className="w-full">
                我已安全保存
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
