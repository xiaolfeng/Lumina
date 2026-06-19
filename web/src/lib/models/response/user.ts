import type { BaseResponse } from './common'

export interface UserInfoResponse {
  username: string
  email: string
  /** 是否已启用生物特征认证 */
  biometric_enabled: boolean
  /** 已注册的生物特征凭据数量 */
  biometric_credential_count: number
}

export type UserInfoResponseWrapper = BaseResponse<UserInfoResponse>

/** 更新个人资料响应 */
export interface UpdateProfileResponse {
  username: string
  email: string
}

export type UpdateProfileResponseWrapper = BaseResponse<UpdateProfileResponse>

/** 更新密码响应 */
export interface UpdatePasswordResponse {
  success: boolean
}

export type UpdatePasswordResponseWrapper = BaseResponse<UpdatePasswordResponse>

/** 生物特征凭据条目 */
export interface BiometricCredentialItem {
  id: string
  device_name: string
  /** Authenticator Attestation GUID */
  aaguid: string
  /** 最后使用时间戳（Unix 秒），从未使用则为 null */
  last_used_at: number | null
  /** 创建时间戳（Unix 秒） */
  created_at: number
}

/** 生物特征凭据列表响应 */
export interface BiometricCredentialListResponse {
  total: number
  items: BiometricCredentialItem[]
}

export type BiometricCredentialListResponseWrapper =
  BaseResponse<BiometricCredentialListResponse>
