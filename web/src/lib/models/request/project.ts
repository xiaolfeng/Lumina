export interface CreateProjectRequest {
  name: string
  alias_name?: string
  match_path?: string[]
  description?: string
  workspace_id: string
}

export interface UpdateProjectRequest {
  workspace_id?: string
  name: string
  alias_name?: string
  match_path?: string[]
  description?: string
}

export interface ProjectListParams {
  page?: number
  size?: number
  workspace_id?: string
  search?: string
}
