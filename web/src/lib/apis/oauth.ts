import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'

export interface OAuthConsentDetail {
  client_name: string
  scope: string
}

export interface OAuthConsentRedirect {
  redirect: string
}

export function getOAuthConsentDetail(
  authorizeId: string,
): Promise<BaseResponse<OAuthConsentDetail>> {
  return apiClient.get('/api/v1/oauth/consent', {
    params: { authorize_id: authorizeId },
  })
}

export function postOAuthConsent(
  authorizeId: string,
  approve: boolean,
): Promise<BaseResponse<OAuthConsentRedirect>> {
  return apiClient.post('/api/v1/oauth/consent', {
    authorize_id: authorizeId,
    approve,
  })
}
