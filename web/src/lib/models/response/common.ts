export interface BaseResponse<T = unknown> {
  code: number
  message: string
  data?: T
  error_message?: string
  context?: string
  output?: string
  overhead?: number
}
