import { useMutation, useQueryClient } from '@tanstack/react-query'
import Cookies from 'js-cookie'
import * as userApi from '#/lib/apis/user'
import type { UpdateProfileResponseWrapper, UpdatePasswordResponseWrapper } from '#/lib/models/response/user'
import type { UpdateProfileRequest, UpdatePasswordRequest } from '#/lib/models/request/user'

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  return useMutation<UpdateProfileResponseWrapper, Error, UpdateProfileRequest>({
    mutationFn: (data) => userApi.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', 'current'] })
    },
  })
}

export function useUpdatePassword() {
  const queryClient = useQueryClient()
  return useMutation<UpdatePasswordResponseWrapper, Error, UpdatePasswordRequest>({
    mutationFn: (data) => userApi.updatePassword(data),
    onSuccess: () => {
      queryClient.clear()
      Cookies.remove('access_token', { path: '/' })
      Cookies.remove('refresh_token', { path: '/' })
      Cookies.remove('expires_at', { path: '/' })
      window.location.href = '/auth/login'
    },
  })
}
