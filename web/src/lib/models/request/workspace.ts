export interface CreateWorkspaceRequest {
  name: string
  slug: string
  description?: string
  icon?: string
}

export interface UpdateWorkspaceRequest {
  name: string
  slug: string
  description?: string
  icon?: string
}

export interface WorkspaceListParams {
  page?: number
  size?: number
}
