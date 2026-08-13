export interface PreviewSessionItem {
  id: string
  project_id: string
  title: string
  hash: string
  status: 'active' | 'deleted'
  created_at: string
  updated_at: string
}

export interface PreviewFileItem {
  id: string
  session_id: string
  filename: string
  mime_type: string
  size: number
  created_at: string
  updated_at: string
}

export interface PreviewSessionDetailResponse {
  session: PreviewSessionItem
  files: PreviewFileItem[]
}

export interface PreviewSessionListResponse {
  items: PreviewSessionItem[]
  total: number
}

export interface PreviewFileDetailResponse extends PreviewFileItem {
  session_hash: string
}
