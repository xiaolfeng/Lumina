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
import { CreatedKeyPanel } from '#/components/mcp/created-key-panel'
import { useCreateApikey } from '#/hooks/useApikey'
import { useMcpEndpoint } from '#/hooks/useMcpEndpoint'
import type { ApikeyCreateResponse } from '#/lib/models/response/apikey'

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateDialog({ open, onOpenChange }: CreateDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [result, setResult] = useState<ApikeyCreateResponse | null>(null)
  const { mcpUrl } = useMcpEndpoint()

  const createMutation = useCreateApikey()

  const handleSubmit = () => {
    if (!name.trim()) return
    createMutation.mutate(
      { name: name.trim(), description: description.trim() || undefined },
      {
        onSuccess: (res) => {
          if (res.data) setResult(res.data)
        },
      },
    )
  }

  const handleClose = () => {
    setName('')
    setDescription('')
    setResult(null)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={result ? undefined : onOpenChange}>
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
              <DialogTitle>创建令牌</DialogTitle>
              <DialogDescription>
                创建一个新的 API 令牌，创建后将显示完整密钥（仅此一次）。
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">名称 *</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="输入令牌名称"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="desc">描述</Label>
                <Input
                  id="desc"
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
                disabled={!name.trim() || createMutation.isPending}
              >
                {createMutation.isPending ? '创建中...' : '创建'}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>密钥已创建</DialogTitle>
              <DialogDescription>
                请立即复制密钥，并按下方模板接入 MCP。关闭后无法再查看完整密钥。
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
