import { useState, useEffect } from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import {
  useAgentModels,
  useUpdateAgentModel,
  useModels,
} from '#/hooks/useLlmConfig'
import type { AgentModelAssignmentItem } from '#/lib/models/response/llm'
import { toast } from 'sonner'

interface AgentModelAssignGroupProps {
  module: string
}

interface RoleDisplay {
  label: string
  description: string
}

const ROLE_DISPLAY_MAP: Partial<Record<string, RoleDisplay>> = {
  'repowiki:coordinator': {
    label: '主控 Agent',
    description: '负责编排决策，协调探索和写作 Agent',
  },
  'repowiki:explore': {
    label: '探索 Agent',
    description: '负责阅读和分析代码库结构',
  },
  'repowiki:write': {
    label: '写作 Agent',
    description: '负责生成结构化的 Wiki 文档',
  },
}

export function AgentModelAssignGroup({
  module,
}: AgentModelAssignGroupProps) {
  const { data: agentData } = useAgentModels(module)
  const { data: modelsData } = useModels()

  const assignments: AgentModelAssignmentItem[] =
    agentData?.data?.assignments ?? []
  const models = modelsData?.data?.items.filter((m) => m.is_active) ?? []

  if (assignments.length === 0) {
    return (
      <div className="rounded-lg border p-4 text-center text-sm text-muted-foreground">
        暂无角色配置
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {assignments.map((assignment) => (
        <AgentModelSelect
          key={assignment.role}
          assignment={assignment}
          models={models}
        />
      ))}
    </div>
  )
}

interface AgentModelSelectProps {
  assignment: AgentModelAssignmentItem
  models: Array<{ id: string; display_name: string }>
}

function AgentModelSelect({
  assignment,
  models,
}: AgentModelSelectProps) {
  const updateMutation = useUpdateAgentModel()
  const [selectedModelId, setSelectedModelId] = useState(
    assignment.model_id ?? '',
  )

  useEffect(() => {
    setSelectedModelId(assignment.model_id ?? '')
  }, [assignment.model_id])

  const handleChange = (value: string) => {
    setSelectedModelId(value)
    updateMutation.mutate(
      { role: assignment.role, data: { model_id: value } },
      {
        onSuccess: () => {
          toast.success(
            `${ROLE_DISPLAY_MAP[assignment.role]?.label ?? assignment.role} 模型分配已更新`,
          )
        },
        onError: (error: Error) => {
          toast.error(error.message || '更新失败')
        },
      },
    )
  }

  const display = ROLE_DISPLAY_MAP[assignment.role]
  const label = display?.label ?? assignment.role
  const description =
    display?.description ?? '为 Agent 角色分配默认使用的模型'

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="space-y-1">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      <div className="w-[240px]">
        <Select value={selectedModelId} onValueChange={handleChange}>
          <SelectTrigger>
            <SelectValue placeholder="选择模型" />
          </SelectTrigger>
          <SelectContent>
            {models.length === 0 && (
              <div className="py-6 text-center text-sm text-muted-foreground">
                暂无可用模型
              </div>
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
