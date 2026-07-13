import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'

// Webhook config response
export interface WebhookConfig {
  url: string
  token: string // masked
  has_secret: boolean
  branches: string[]
  provider_guide: Record<string, string>
}

// Regenerate response (one-time plaintext)
export interface RegenerateWebhookResponse {
  token: string
  secret: string
  url: string
}

// Webhook event
export interface WebhookEvent {
  id: string
  config_id?: string
  provider: string
  event_type: string
  branch?: string
  commit_before?: string
  commit_after?: string
  changed_count: number
  status: string
  reason?: string
  version_id?: string
  response_code: number
  received_at: string
  processed_at?: string
}

export interface WebhookEventList {
  total: number
  items: WebhookEvent[]
}

export async function getWebhookConfig(configId: string): Promise<BaseResponse<WebhookConfig>> {
  return apiClient.get(`/api/v1/repowiki/configs/${configId}/webhook`)
}

export async function updateWebhookBranches(
  configId: string,
  branches: string[],
): Promise<BaseResponse> {
  return apiClient.put(`/api/v1/repowiki/configs/${configId}/webhook/branches`, { branches })
}

export async function regenerateWebhook(
  configId: string,
): Promise<BaseResponse<RegenerateWebhookResponse>> {
  return apiClient.post(`/api/v1/repowiki/configs/${configId}/webhook/regenerate`)
}

export async function listWebhookEvents(
  configId: string,
  page: number,
  size: number,
): Promise<BaseResponse<WebhookEventList>> {
  return apiClient.get(`/api/v1/repowiki/configs/${configId}/webhook/events`, {
    params: { page, size },
  })
}
