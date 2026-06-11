export interface CreateProjectRequest {
  name: string
  alias_name?: string
  match_path?: string[]
  description?: string
}

export interface UpdateProjectRequest {
  name: string
  alias_name?: string
  match_path?: string[]
  description?: string
}

export interface ProjectListParams {
  page?: number
  size?: number
}
