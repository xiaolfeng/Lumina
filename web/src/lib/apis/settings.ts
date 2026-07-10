import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'

export interface SettingItem {
  key: string
  value: string
  type: string
  description: string
}

export interface CategorySettings {
  category: string
  items: SettingItem[]
}

export function getSettings(
  category: string,
): Promise<BaseResponse<CategorySettings>> {
  return apiClient.get(`/api/v1/settings/${category}`)
}

export function updateSettings(
  category: string,
  items: { key: string; value: string }[],
): Promise<BaseResponse> {
  return apiClient.put(`/api/v1/settings/${category}`, { items })
}
