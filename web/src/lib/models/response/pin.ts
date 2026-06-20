export interface PinItem {
  id: string
  title: string
  content: string
  category: string
  status: string
  priority: string
  from_project_id: string
  to_project_id: string
  consumed_at: string
  created_at: string
  updated_at: string
}

export interface PinListResponse {
  items: PinItem[]
  total: number
}
