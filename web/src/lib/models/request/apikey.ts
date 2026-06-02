export interface ApikeyCreateRequest {
  name: string
  description?: string
  expires_at?: string
}

export interface ApikeyUpdateRequest {
  name?: string
  description?: string
  expires_at?: string
  is_active?: boolean
}

export interface ApikeyListParams {
  page?: number
  size?: number
}
