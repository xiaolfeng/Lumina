export interface SessionListParams {
  page?: number
  size?: number
  status?: 'active' | 'expired' | 'deleted' | ''
  type?: 'temporary' | 'permanent' | ''
}

export interface UpdateQaConfigRequest {
  session_ttl?: number
  runtime_domain?: string
}
