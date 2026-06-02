export interface ApikeyItem {
  id: string
  name: string
  key: string
  key_prefix: string
  description: string
  expires_at: string | null
  is_active: boolean
  last_used_at: string | null
  created_at: string
}

export interface ApikeyCreateResponse {
  id: string
  name: string
  key: string
  key_prefix: string
  description: string
  expires_at: string | null
  is_active: boolean
  created_at: string
}

export interface ApikeyDetailResponse {
  id: string
  name: string
  key: string
  key_prefix: string
  description: string
  expires_at: string | null
  is_active: boolean
  last_used_at: string | null
  created_at: string
}

export interface ApikeyResetResponse {
  id: string
  name: string
  key: string
  key_prefix: string
  description: string
  expires_at: string | null
  is_active: boolean
  created_at: string
}
