import type { BaseResponse } from './common'

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export type TokenResponseWrapper = BaseResponse<TokenResponse>

export interface StatusResponse {
  is_initial: boolean
}

export type StatusResponseWrapper = BaseResponse<StatusResponse>
