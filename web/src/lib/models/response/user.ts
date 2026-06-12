import type { BaseResponse } from './common'

export interface UserInfoResponse {
  username: string
  email: string
}

export type UserInfoResponseWrapper = BaseResponse<UserInfoResponse>
