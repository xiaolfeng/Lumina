import { apiClient } from './client'
import type {
  AvailabilityResponseWrapper,
  RegisterStartResponseWrapper,
  RegisterFinishResponseWrapper,
  LoginStartResponseWrapper,
  LoginFinishResponseWrapper,
} from '../models/response/biometric'
import type { RegisterStartRequest, RegisterFinishRequest, LoginFinishRequest } from '../models/request/biometric'

/**
 * 获取生物特征登录可用性（公开接口）
 * 检测当前设备是否支持平台认证器
 */
export async function getAvailability(): Promise<AvailabilityResponseWrapper> {
  return apiClient.get('/api/v1/auth/biometric/availability')
}

/**
 * 注册开始（需认证）
 * 向服务端请求 WebAuthn 创建选项，用于注册新的生物特征凭据
 */
export async function registerStart(
  data: RegisterStartRequest,
): Promise<RegisterStartResponseWrapper> {
  return apiClient.post('/api/v1/auth/biometric/register/start', data)
}

/**
 * 注册完成（需认证）
 * 将创建的 PublicKeyCredential 提交到服务端完成绑定
 */
export async function registerFinish(
  data: RegisterFinishRequest,
): Promise<RegisterFinishResponseWrapper> {
  return apiClient.post('/api/v1/auth/biometric/register/finish', data)
}

/**
 * 登录开始（公开）
 * 向服务端请求 WebAuthn 获取选项，用于生物特征登录
 */
export async function loginStart(): Promise<LoginStartResponseWrapper> {
  return apiClient.post('/api/v1/auth/biometric/login/start')
}

/**
 * 登录完成（公开）
 * 将验证的 PublicKeyCredential 提交到服务端换取 Token
 */
export async function loginFinish(
  data: LoginFinishRequest,
): Promise<LoginFinishResponseWrapper> {
  return apiClient.post('/api/v1/auth/biometric/login/finish', data)
}
