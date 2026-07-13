import { useState } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Badge } from '#/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { useWebhookConfig, useUpdateWebhookBranches } from '#/hooks/useWebhook'
import { X, Plus } from 'lucide-react'
import { toast } from 'sonner'

interface WebhookBranchesProps {
  configId: string
}

export function WebhookBranches({ configId }: WebhookBranchesProps) {
  const { data, isLoading } = useWebhookConfig(configId)
  const updateMutation = useUpdateWebhookBranches(configId)
  const [newBranch, setNewBranch] = useState('')

  const branches = data?.data?.branches ?? []

  const handleAdd = () => {
    const trimmed = newBranch.trim()
    if (!trimmed) return
    if (branches.includes(trimmed)) {
      toast.error('该分支已存在')
      return
    }
    const updated = [...branches, trimmed]
    updateMutation.mutate(updated, {
      onSuccess: () => {
        setNewBranch('')
        toast.success('分支已添加')
      },
      onError: (error: Error) => {
        toast.error(error.message || '添加失败')
      },
    })
  }

  const handleRemove = (branch: string) => {
    const updated = branches.filter((b) => b !== branch)
    updateMutation.mutate(updated, {
      onSuccess: () => {
        toast.success('分支已移除')
      },
      onError: (error: Error) => {
        toast.error(error.message || '移除失败')
      },
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAdd()
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>监听分支</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-8 animate-pulse rounded bg-muted" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>监听分支</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Add branch */}
        <div className="flex gap-2">
          <Input
            value={newBranch}
            onChange={(e) => setNewBranch(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入分支名称（如 main）"
            disabled={updateMutation.isPending}
          />
          <Button
            onClick={handleAdd}
            disabled={!newBranch.trim() || updateMutation.isPending}
            className="gap-2"
          >
            <Plus className="size-4" />
            添加
          </Button>
        </div>

        {/* Branch list */}
        {branches.length === 0 ? (
          <div className="rounded-md border border-dashed p-6 text-center">
            <p className="text-muted-foreground text-sm">
              未配置监听分支，Webhook 不会触发分析
            </p>
          </div>
        ) : (
          <div className="flex flex-wrap gap-2">
            {branches.map((branch) => (
              <Badge
                key={branch}
                variant="secondary"
                className="gap-1 px-2 py-1 text-sm"
              >
                {branch}
                <button
                  onClick={() => handleRemove(branch)}
                  disabled={updateMutation.isPending}
                  className="ml-1 rounded-full p-0.5 hover:bg-muted-foreground/20 transition-colors disabled:opacity-50"
                  aria-label={`移除分支 ${branch}`}
                >
                  <X className="size-3" />
                </button>
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
