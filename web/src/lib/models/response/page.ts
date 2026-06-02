export interface PageResponse<T> {
  current_page: number
  total_pages: number
  total_items: number
  size: number
  items: T[]
}
