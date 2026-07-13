export interface SessionListParams {
  page?: number
  size?: number
  status?: 'active' | 'expired' | 'deleted' | ''
  type?: 'temporary' | 'permanent' | ''
  hash?: string
}

export interface CreateSessionRequest {
  project_id: string
  title: string
  agent: string
  type: 'temporary' | 'permanent'
  link_domain?: string
}

export interface UpdateQaConfigRequest {
  session_ttl?: number
  runtime_domain?: string
  poll_slice?: number
  max_retries?: number
}
