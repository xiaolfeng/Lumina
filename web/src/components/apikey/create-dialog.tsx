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
import { useCreateApikey } from '#/hooks/useApikey'
import type { ApikeyCreateResponse } from '#/lib/models/response/apikey'
import { Copy, Check } from 'lucide-react'
import { toast } from 'sonner'

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateDialog({ open, onOpenChange }: CreateDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [result, setResult] = useState<ApikeyCreateResponse | null>(null)
  const [copied, setCopied] = useState(false)

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

  const handleCopy = async () => {
    if (!result?.key) return
    await navigator.clipboard.writeText(result.key)
    setCopied(true)
    toast.success('已复制到剪贴板')
    setTimeout(() => setCopied(false), 2000)
  }

  const handleClose = () => {
    setName('')
    setDescription('')
    setResult(null)
    setCopied(false)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={result ? undefined : onOpenChange}>
      <DialogContent
        onPointerDownOutside={result ? (e) => e.preventDefault() : undefined}
      >
        {!result ? (
          <>
            <DialogHeader>
              <DialogTitle>创建令牌</DialogTitle>
              <DialogDescription>
                创建一个新的 API
                令牌，创建后将显示完整密钥（仅此一次）。
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
                请立即复制并安全保存此密钥，关闭后将无法再次查看。
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
