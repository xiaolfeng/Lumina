import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  UserInfoResponseWrapper,
  UpdateProfileResponseWrapper,
  UpdatePasswordResponseWrapper,
  BiometricCredentialListResponseWrapper,
} from '../models/response/user'
import type { UpdateProfileRequest, UpdatePasswordRequest } from '../models/request/user'

/** 获取当前用户信息 */
export function getCurrentUser(): Promise<UserInfoResponseWrapper> {
  return apiClient.get('/api/v1/user/current')
}

/**
 * 更新个人资料（用户名 / 邮箱）
 */
export async function updateProfile(
  data: UpdateProfileRequest,
): Promise<UpdateProfileResponseWrapper> {
  return apiClient.put('/api/v1/user/profile', data)
}

/**
 * 修改登录密码
 */
export async function updatePassword(
  data: UpdatePasswordRequest,
): Promise<UpdatePasswordResponseWrapper> {
  return apiClient.put('/api/v1/user/password', data)
}

/**
 * 获取已注册的生物特征凭据列表
 */
export async function getBiometricCredentials(): Promise<BiometricCredentialListResponseWrapper> {
  return apiClient.get('/api/v1/user/biometric/credentials')
}

/**
 * 删除指定的生物特征凭据
 * @param id - 凭据 ID
 */
export async function deleteBiometricCredential(
  id: string,
): Promise<BaseResponse<null>> {
  return apiClient.delete(`/api/v1/user/biometric/credentials/${id}`)
}