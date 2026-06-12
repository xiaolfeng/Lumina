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
