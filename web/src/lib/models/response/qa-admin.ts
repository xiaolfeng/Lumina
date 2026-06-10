export interface SessionItem {
  id: string
  title: string
  agent: string
  type: 'temporary' | 'permanent'
  status: 'active' | 'expired' | 'deleted'
  online_devices: number
  expires_at: string
  created_at: string
  updated_at: string
}

export interface SessionListResponse {
  items: SessionItem[]
  total: number
}

export interface QuestionSummary {
  id: string
  type: string
  title: string
  status: 'pending' | 'answered' | 'skipped'
  created_at: string
  answered_at: string
}

export interface SessionDetailResponse {
  id: string
  title: string
  agent: string
  type: 'temporary' | 'permanent'
  status: 'active' | 'expired' | 'deleted'
  online_devices: number
  expires_at: string
  created_at: string
  updated_at: string
  questions: QuestionSummary[]
}

export interface SupplementItem {
  id: string
  target_type: 'question' | 'option'
  target_id: string
  content_type: 'markdown' | 'html'
  content: string
  created_at: string
  updated_at: string
}

export interface QuestionDetailResponse {
  id: string
  session_id: string
  type: string
  title: string
  description: string
  options: any
  config: any
  batch: any
  group_label: string
  status: 'pending' | 'answered' | 'skipped'
  answer: any
  supplements: SupplementItem[]
  created_at: string
  answered_at: string
}

export interface QaConfigResponse {
  session_ttl: number
  runtime_domain: string
}
