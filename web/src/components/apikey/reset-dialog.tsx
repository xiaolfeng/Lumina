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
import { CreatedKeyPanel } from '#/components/mcp/created-key-panel'
import { useResetApikey } from '#/hooks/useApikey'
import { useMcpEndpoint } from '#/hooks/useMcpEndpoint'
import type {
  ApikeyItem,
  ApikeyResetResponse,
} from '#/lib/models/response/apikey'
import { AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

interface ResetDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApikeyItem | null
}

export function ResetDialog({ open, onOpenChange, item }: ResetDialogProps) {
  const [result, setResult] = useState<ApikeyResetResponse | null>(null)
  const { mcpUrl } = useMcpEndpoint()

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

  const handleClose = () => {
    setResult(null)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={result ? undefined : handleClose}>
      <DialogContent
        className={
          result ? 'sm:max-w-2xl max-h-[85vh] overflow-y-auto' : undefined
        }
        showCloseButton={!result}
        onPointerDownOutside={result ? (e) => e.preventDefault() : undefined}
      >
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
                <p>重置后旧密钥将立即失效，所有使用旧密钥的服务将无法访问。</p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={handleClose}>
                取消
              </Button>
              <Button onClick={handleReset} disabled={resetMutation.isPending}>
                {resetMutation.isPending ? '重置中...' : '确认重置'}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>密钥已重置</DialogTitle>
              <DialogDescription>
                请立即复制新密钥，并按下方模板更新客户端配置。关闭后无法再查看。
              </DialogDescription>
            </DialogHeader>
            <CreatedKeyPanel
              apiKey={result.key}
              mcpUrl={mcpUrl}
              onDismiss={handleClose}
            />
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
