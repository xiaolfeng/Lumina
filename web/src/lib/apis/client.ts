import axios from 'axios'
import Cookies from 'js-cookie'
import type { BaseResponse } from '../models/response/common'

export const apiClient = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = Cookies.get('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use((response) => {
  const data = response.data
  if (data && typeof data === 'object' && 'code' in data) {
    const baseData = data as BaseResponse
    if (baseData.code !== 200) {
      throw new Error(baseData.message)
    }
  }
  return response.data
})
