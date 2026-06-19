import type { BaseResponse } from './common'
import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
} from '../../webauthn/helpers'

/** 生物特征登录可用性响应 */
export interface AvailabilityResponse {
  is_available: boolean
}

export type AvailabilityResponseWrapper = BaseResponse<AvailabilityResponse>

/** 注册开始响应 — 包含 WebAuthn 创建选项 */
export interface RegisterStartResponse {
  session_token: string
  options: PublicKeyCredentialCreationOptionsJSON
}

export type RegisterStartResponseWrapper = BaseResponse<RegisterStartResponse>

/** 注册完成响应 */
export interface RegisterFinishResponse {
  success: boolean
  message: string
}

export type RegisterFinishResponseWrapper = BaseResponse<RegisterFinishResponse>

/** 登录开始响应 — 包含 WebAuthn 获取选项 */
export interface LoginStartResponse {
  session_token: string
  options: PublicKeyCredentialRequestOptionsJSON
}

export type LoginStartResponseWrapper = BaseResponse<LoginStartResponse>

/** 登录完成响应 — 返回 Token（与 TokenResponse 结构对齐） */
export interface LoginFinishResponse {
  access_token: string
  refresh_token: string
  expires_at: number
}

export type LoginFinishResponseWrapper = BaseResponse<LoginFinishResponse>
