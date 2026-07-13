// LLM Provider 请求
export interface CreateProviderRequest {
  name: string
  protocol: string
  base_url: string
  api_key: string
  description: string
}

export interface UpdateProviderRequest {
  name?: string
  protocol?: string
  base_url?: string
  api_key?: string
  is_active?: boolean
  description?: string
}

// LLM Model 请求
export interface CreateModelRequest {
  provider_id: string
  model_name: string
  display_name: string
  max_tokens: number
  context_window: number
  temperature: number
  description: string
}

export interface UpdateModelRequest {
  model_name?: string
  display_name?: string
  max_tokens?: number
  context_window?: number
  temperature?: number
  is_active?: boolean
  description?: string
}

// Agent 请求
export interface UpdateAgentModelRequest {
  model_id: string
}
