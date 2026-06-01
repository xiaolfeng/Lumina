import { useEffect, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import * as api from '#/lib/apis/auth'
import type {
  InitializeRequest,
  LoginRequest,
  RefreshRequest,
} from '#/lib/models/request/auth'
import type { TokenResponse, StatusResponse } from '#/lib/models/response/auth'
import type { BaseResponse } from '#/lib/models/response/common'

// ── Cookie helpers ──

export function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[2]) : null
}

export function setCookie(name: string, value: string, maxAge: number): void {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAge}; SameSite=Lax`
}

export function deleteCookie(name: string): void {
  document.cookie = `${name}=; path=/; max-age=0; SameSite=Lax`
}

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
        setCookie('access_token', tokenData.access_token, tokenData.expires_in)
        setCookie(
          'refresh_token',
          tokenData.refresh_token,
          THIRTY_DAYS_IN_SECONDS,
        )
        setCookie(
          'expires_at',
          String(Date.now() + tokenData.expires_in * 1000),
          tokenData.expires_in,
        )
      }
    },
  })
}

export function useLogout() {
  return useMutation<BaseResponse, Error, RefreshRequest>({
    mutationFn: api.logout,
    onSuccess: () => {
      deleteCookie('access_token')
      deleteCookie('refresh_token')
      deleteCookie('expires_at')
    },
  })
}

export function useRefresh() {
  return useMutation<BaseResponse<TokenResponse>, Error, RefreshRequest>({
    mutationFn: api.refresh,
    onSuccess: (response) => {
      const tokenData = response.data
      if (tokenData) {
        setCookie('access_token', tokenData.access_token, tokenData.expires_in)
        setCookie(
          'refresh_token',
          tokenData.refresh_token,
          THIRTY_DAYS_IN_SECONDS,
        )
        setCookie(
          'expires_at',
          String(Date.now() + tokenData.expires_in * 1000),
          tokenData.expires_in,
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

// ── Composite hook ──

export function useAuth() {
  const initialize = useInitialize()
  const login = useLogin()
  const logout = useLogout()
  const refresh = useRefresh()
  const status = useStatus()

  const isAuthenticated = getCookie('access_token') !== null

  const refreshRef = useRef(refresh)
  refreshRef.current = refresh

  useEffect(() => {
    const intervalId = setInterval(() => {
      const expiresAt = getCookie('expires_at')
      if (!expiresAt) return

      const expiresAtMs = Number(expiresAt)
      if (Number.isNaN(expiresAtMs)) return

      const fiveMinutes = 5 * 60 * 1000
      if (expiresAtMs - Date.now() < fiveMinutes) {
        const refreshToken = getCookie('refresh_token')
        if (refreshToken && !refreshRef.current.isPending) {
          refreshRef.current.mutate({ refresh_token: refreshToken })
        }
      }
    }, 30 * 1000)

    return () => clearInterval(intervalId)
  }, [])

  return {
    isAuthenticated,
    initialize,
    login,
    logout,
    refresh,
    status,
  }
}
