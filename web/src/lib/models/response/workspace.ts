export interface WorkspaceItem {
  id: string
  name: string
  slug: string
  description: string
  icon: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface WorkspaceListResponse {
  items: WorkspaceItem[]
  total: number
}

export interface DeleteWorkspaceResponse {
  moved_project_count: number
}
