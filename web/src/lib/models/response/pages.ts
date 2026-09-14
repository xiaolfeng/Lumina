export interface PageItem {
  id: string
  project_id: string
  project_name: string
  slug: string
  title: string
  description: string
  status: 'published' | 'archived'
  access_mode: 'public' | 'password'
  latest_version_id: string
  latest_version: string
  page_url: string
  created_at: string
  updated_at: string
}

export interface PageVersionItem {
  id: string
  page_id: string
  version: string
  changelog: string
  source_session_id?: string
  base_version_id?: string
  entry_filename: string
  file_count: number
  total_size: number
  created_by: string
  is_active: boolean
  created_at: string
}

export interface PageFileItem {
  id: string
  version_id: string
  filename: string
  mime_type: string
  size: number
  created_at: string
  updated_at: string
}

export interface PageListResponse {
  items: PageItem[]
  total: number
}

export interface PageVersionListResponse {
  items: PageVersionItem[]
}

export interface PagePublicMetaResponse {
  page: PageItem
  version: PageVersionItem
  files: PageFileItem[]
}

export interface PageAuthCheckResponse {
  authenticated: boolean
  password_required: boolean
}

export interface PromoteConflict {
  current_version: string
  current_version_id: string
  current_changelog: string
  current_created_by: string
  current_created_at: string
  source_version_id: string
  message: string
}

export interface PromoteSessionResponse {
  conflict?: PromoteConflict
  page?: PageItem
  version?: PageVersionItem
}

export interface ForkPageResponse {
  session: {
    id: string
    hash: string
    title: string
  }
  preview_url: string
}
