/**
 * Wiki Reader 独立 API 客户端
 *
 * 与 web/ 的 apiClient 完全隔离：
 * - 使用 Cookie 授权（非 Bearer Token）
 * - withCredentials: true 发送/接收 Cookie
 * - 无 Token 刷新逻辑
 */
import axios from 'axios'
import type { AxiosError, AxiosResponse } from 'axios'

// ── 类型定义 ──

/** auth-check 接口响应数据 */
export interface AuthCheckResponse {
  password_required: boolean
  authenticated: boolean
  wiki_name?: string
  wiki_description?: string
}

/** auth 接口响应数据 */
export interface AuthResponse {
  success: boolean
  message?: string
}

/** getPage 接口响应数据（Task 25 定义） */
export interface PageResponse {
  path: string
  content: string
  title?: string
}

/** getManifest 接口响应数据（与后端 WikiManifestResponse 对齐） */
export interface WikiNavItem {
  title: string
  path: string
  children?: WikiNavItem[]
}

export interface ManifestResponse {
  navigation: WikiNavItem[]
  home: string
  language: string
  project_name: string
}

// ── Axios 实例 ──

export const wikiApi = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ── 响应拦截器：统一错误处理 ──

wikiApi.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    // 业务错误码处理（与后端 BaseResponse 格式对齐）
    if (
      data &&
      typeof data === 'object' &&
      'code' in data &&
      data.code !== 200
    ) {
      return Promise.reject(
        new Error(
          data.error_message ?? data.message ?? `请求失败 (code: ${data.code})`,
        ),
      )
    }
    return data
  },
  (error: AxiosError) => {
    if (error.response) {
      const status = error.response.status
      const data = error.response.data
      // HTTP 错误 + 业务错误码
      let msg: string | undefined
      if (typeof data === 'object' && data !== null) {
        msg =
          (data as { error_message?: string }).error_message ??
          (data as { message?: string }).message
      }
      if (!msg) {
        msg = `HTTP ${status} 错误`
      }
      return Promise.reject(new Error(msg))
    }
    if (error.request) {
      return Promise.reject(new Error('网络连接失败，请检查网络'))
    }
    return Promise.reject(new Error(error.message || '未知错误'))
  },
)

// ── Wiki Reader API 封装 ──

export const wikiReaderApi = {
  /** 检查 Wiki 授权状态 */
  checkAuth: (wikiId: string): Promise<AuthCheckResponse> =>
    wikiApi.get(`/wiki/${wikiId}/auth-check`).then((res) => res.data),

  /** 密码验证 */
  auth: (wikiId: string, password: string): Promise<AuthResponse> =>
    wikiApi.post(`/wiki/${wikiId}/auth`, { password }).then((res) => res.data),

  /** 获取页面内容（Task 25 使用） */
  getPage: (wikiId: string, path: string): Promise<PageResponse> =>
    wikiApi.get(`/wiki/${wikiId}/page/${path}`).then((res) => res.data),

  /** 获取导航清单（Task 26 使用） */
  getManifest: (wikiId: string): Promise<ManifestResponse> =>
    wikiApi.get(`/wiki/${wikiId}/manifest`).then((res) => res.data),
}
