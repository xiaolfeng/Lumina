import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type { DashboardOverview } from '../models/response/dashboard'

export function getDashboardOverview(): Promise<BaseResponse<DashboardOverview>> {
  return apiClient.get('/api/v1/dashboard/overview')
}
