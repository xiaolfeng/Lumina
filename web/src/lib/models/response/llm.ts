// LLM Provider 响应
export interface Provider {
  id: string
  name: string
  protocol: string
  base_url: string
  has_key: boolean
  is_active: boolean
  description: string
  created_at: string
  updated_at?: string
}

export interface ProviderListResponse {
  items: Provider[]
  total_items: number
  current_page: number
  total_pages: number
  size: number
}

// LLM Model 响应
export interface Model {
  id: string
  provider_id: string
  model_name: string
  display_name: string
  max_tokens: number
  temperature: number
  is_active: boolean
  description: string
  created_at: string
  updated_at?: string
}

export interface ModelListResponse {
  items: Model[]
  total_items: number
  current_page: number
  total_pages: number
  size: number
}

// Agent 响应
export interface AgentModelAssignment {
  role: string
  model_id: string | null
}

// Agent 模型分配项（多角色批量查询）
export interface AgentModelAssignmentItem {
  role: string
  model_id: string | null
  model_name: string | null
}

// Agent 模型分配批量响应
export interface AgentModelAssignmentsResponse {
  module: string
  assignments: AgentModelAssignmentItem[]
}
