import { useEffect, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import Cookies from 'js-cookie'
import * as api from '#/lib/apis/auth'
import { getCurrentUser } from '#/lib/apis/user'
import type {
  InitializeRequest,
  LoginRequest,
  RefreshRequest,
} from '#/lib/models/request/auth'
import type { TokenResponse, StatusResponse } from '#/lib/models/response/auth'
import type { UserInfoResponse } from '#/lib/models/response/user'
import * as biometricApi from '#/lib/apis/biometric'
import { getCredential, bufferToBase64url } from '#/lib/webauthn/helpers'
import type { AvailabilityResponse, LoginFinishResponse } from '#/lib/models/response/biometric'
import type { BaseResponse } from '#/lib/models/response/common'

export const getCookie = Cookies.get

const THIRTY_DAYS_IN_SECONDS = 30 * 24 * 60 * 60

// ── Mutations ──

export function useInitialize() {
  return useMutation<BaseResponse, Error, InitializeRequest>({
    mutationFn: api.initialize,
  })
}

export function useLogin() {
  return useMutation<BaseResponse<TokenResponse>, Error, LoginRequest>({
    mutationFn: api.login,
    onSuccess: (response) => {
      const tokenData = response.data
      if (tokenData) {
        Cookies.set('access_token', tokenData.access_token, {
          expires: tokenData.expires_in / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        Cookies.set('refresh_token', tokenData.refresh_token, {
          expires: THIRTY_DAYS_IN_SECONDS / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        Cookies.set(
          'expires_at',
          String(Date.now() + tokenData.expires_in * 1000),
          {
            expires: tokenData.expires_in / 86400,
            path: '/',
            sameSite: 'Lax',
          },
        )
      }
    },
  })
}

export function useLogout() {
  return useMutation<BaseResponse, Error, RefreshRequest>({
    mutationFn: api.logout,
    onSuccess: () => {
      Cookies.remove('access_token', { path: '/' })
      Cookies.remove('refresh_token', { path: '/' })
      Cookies.remove('expires_at', { path: '/' })
    },
  })
}

export function useRefresh() {
  return useMutation<BaseResponse<TokenResponse>, Error, RefreshRequest>({
    mutationFn: api.refresh,
    onSuccess: (response) => {
      const tokenData = response.data
      if (tokenData) {
        Cookies.set('access_token', tokenData.access_token, {
          expires: tokenData.expires_in / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        Cookies.set('refresh_token', tokenData.refresh_token, {
          expires: THIRTY_DAYS_IN_SECONDS / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        Cookies.set(
          'expires_at',
          String(Date.now() + tokenData.expires_in * 1000),
          {
            expires: tokenData.expires_in / 86400,
            path: '/',
            sameSite: 'Lax',
          },
        )
      }
    },
  })
}

// ── Queries ──

export function useStatus() {
  return useQuery<BaseResponse<StatusResponse>, Error>({
    queryKey: ['auth', 'status'],
    queryFn: api.getStatus,
    staleTime: 0,
    enabled: true,
  })
}

export function useCurrentUser() {
  const hasToken = !!Cookies.get('access_token') || !!Cookies.get('refresh_token')
  return useQuery<BaseResponse<UserInfoResponse>, Error>({
    queryKey: ['user', 'current'],
    queryFn: getCurrentUser,
    staleTime: 5 * 60 * 1000,
    enabled: hasToken,
    retry: false,
  })
}

// ── Biometric Helpers ──

/**
 * 将 PublicKeyCredential 序列化为可提交的 JSON 对象
 * 遵循 WebAuthn 规范的 JSON 序列化格式
 */
function serializeCredentialForRequest(credential: PublicKeyCredential): unknown {
  const response = credential.response as AuthenticatorAssertionResponse
  return {
    id: credential.id,
    rawId: credential.id, // base64url
    type: credential.type,
    authenticatorAttachment: credential.authenticatorAttachment,
    response: {
      authenticatorData: bufferToBase64url(response.authenticatorData),
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
      signature: bufferToBase64url(response.signature),
      userHandle: response.userHandle
        ? bufferToBase64url(response.userHandle)
        : null,
    },
    clientExtensionResults: {},
  }
}

// ── Biometric Hooks ──

/**
 * 查询生物特征登录可用性（公开接口，无需认证）
 * 用于登录页判断是否显示生物特征登录按钮
 */
export function useBiometricAvailability() {
  return useQuery<BaseResponse<AvailabilityResponse>, Error>({
    queryKey: ['biometric', 'availability'],
    queryFn: biometricApi.getAvailability,
    staleTime: 60 * 1000, // 1 分钟内不重复请求
    retry: false,
  })
}

/**
 * 生物特征登录（公开接口）
 * 内部流程：loginStart → navigator.credentials.get → loginFinish → 写入 Cookie
 */
export function useBiometricLogin() {
  return useMutation<BaseResponse<LoginFinishResponse>, Error, void>({
    mutationFn: async () => {
      // 1. 登录开始：获取 WebAuthn 请求选项
      const startResp = await biometricApi.loginStart()
      const startData = startResp.data
      if (!startData) {
        throw new Error('登录开始失败：未返回数据')
      }

      // 2. 浏览器 WebAuthn 认证
      const credential = await getCredential(startData.options)
      if (!credential) {
        throw new Error('生物特征认证已取消')
      }

      // 3. 序列化 credential 为 JSON（WebAuthn 规范的 JSON 序列化）
      const credentialJSON = serializeCredentialForRequest(credential)

      // 4. 登录完成：提交凭证换取 Token
      return biometricApi.loginFinish({
        session_token: startData.session_token,
        credential: credentialJSON,
      })
    },
    onSuccess: (response) => {
      const tokenData = response.data
      if (tokenData) {
        // 复用 useLogin 的 Cookie 写入逻辑
        Cookies.set('access_token', tokenData.access_token, {
          expires: THIRTY_DAYS_IN_SECONDS / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        Cookies.set('refresh_token', tokenData.refresh_token, {
          expires: THIRTY_DAYS_IN_SECONDS / 86400,
          path: '/',
          sameSite: 'Lax',
        })
        // ⚠️ 后端 LoginFinishResponse 返回的是 expires_at（时间戳秒），没有 expires_in
        // 需要计算 expiresIn 用于 Cookie 过期
        const expiresIn = Math.max(
          0,
          Math.floor((tokenData.expires_at - Date.now() / 1000)),
        )
        Cookies.set('expires_at', String(tokenData.expires_at * 1000), {
          expires: expiresIn / 86400,
          path: '/',
          sameSite: 'Lax',
        })
      }
    },
  })
}

// ── Composite hook ──

export function useAuth() {
  const initialize = useInitialize()
  const login = useLogin()
  const logout = useLogout()
  const refresh = useRefresh()
  const status = useStatus()
  const currentUser = useCurrentUser()

  const isAuthenticated = !!Cookies.get('access_token') || !!Cookies.get('refresh_token')

  const refreshRef = useRef(refresh)
  refreshRef.current = refresh

  useEffect(() => {
    const tryRefreshIfNeeded = () => {
      const expiresAt = Cookies.get('expires_at')
      if (!expiresAt) return

      const expiresAtMs = Number(expiresAt)
      if (Number.isNaN(expiresAtMs)) return

      const fiveMinutes = 5 * 60 * 1000
      if (expiresAtMs - Date.now() < fiveMinutes) {
        const refreshToken = Cookies.get('refresh_token')
        if (refreshToken && !refreshRef.current.isPending) {
          refreshRef.current.mutate({ refresh_token: refreshToken })
        }
      }
    }

    const intervalId = setInterval(tryRefreshIfNeeded, 30 * 1000)

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        tryRefreshIfNeeded()
      }
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      clearInterval(intervalId)
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [])

  return {
    isAuthenticated,
    currentUser,
    initialize,
    login,
    logout,
    refresh,
    status,
  }
}
