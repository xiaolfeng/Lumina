import { useState, useEffect } from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import {
  useAgentModel,
  useUpdateAgentModel,
  useModels,
} from '#/hooks/useLlmConfig'
import { toast } from 'sonner'

interface AgentModelAssignProps {
  role: string
}

const ROLE_DISPLAY_MAP: Record<string, string> = {
  repowiki: 'RepoWiki 分析',
}

export function AgentModelAssign({ role }: AgentModelAssignProps) {
  const { data: agentData } = useAgentModel(role)
  const { data: modelsData } = useModels()
  const updateMutation = useUpdateAgentModel()

  const [selectedModelId, setSelectedModelId] = useState('')

  const models =
    modelsData?.data?.list.filter((m) => m.is_active) || []

  useEffect(() => {
    if (agentData?.data?.model_id) {
      setSelectedModelId(agentData.data.model_id)
    }
  }, [agentData])

  const handleChange = (value: string) => {
    setSelectedModelId(value)
    updateMutation.mutate(
      { role, data: { model_id: value } },
      {
        onSuccess: () => {
          toast.success('模型分配已更新')
        },
        onError: (error: Error) => {
          toast.error(error.message || '更新失败')
        },
      },
    )
  }

  const displayRole = ROLE_DISPLAY_MAP[role] || role

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="space-y-1">
        <p className="text-sm font-medium">{displayRole}</p>
        <p className="text-xs text-muted-foreground">
          为 Agent 角色分配默认使用的模型
        </p>
      </div>
      <div className="w-[240px]">
        <Select value={selectedModelId} onValueChange={handleChange}>
          <SelectTrigger>
            <SelectValue placeholder="选择模型" />
          </SelectTrigger>
          <SelectContent>
            {models.length === 0 && (
              <SelectItem value="" disabled>
                暂无可用模型
              </SelectItem>
            )}
            {models.map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.display_name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
