export interface ProjectItem {
  id: string
  name: string
  alias_name: string
  match_path: string[]
  description: string
  created_at: string
  updated_at: string
}

export interface ProjectListResponse {
  items: ProjectItem[]
  total: number
}
