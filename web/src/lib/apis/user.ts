import { apiClient } from './client'
import type { UserInfoResponseWrapper } from '../models/response/user'

export function getCurrentUser(): Promise<UserInfoResponseWrapper> {
  return apiClient.get('/api/v1/user/current')
}
