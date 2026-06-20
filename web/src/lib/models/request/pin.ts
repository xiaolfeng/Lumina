export interface CreatePinRequest {
  title: string
  content: string
  category?: string
  priority: string
  from_project_id?: string
  to_project_id: string
}

export interface UpdatePinRequest {
  priority?: string
  category?: string
}

export interface PinListParams {
  to_project_id?: string
  from_project_id?: string
  status?: string
  category?: string
  priority?: string
  page?: number
  size?: number
}
