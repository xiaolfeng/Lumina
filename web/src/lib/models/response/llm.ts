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
  list: Provider[]
  total: number
  page: number
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
  list: Model[]
  total: number
  page: number
  size: number
}

// Agent 响应
export interface AgentModelAssignment {
  role: string
  model_id: string | null
}
