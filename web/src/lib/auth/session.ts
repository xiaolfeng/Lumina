const REFRESH_LEAD_MS = 5 * 60 * 1000

const SESSION_NEUTRAL_URLS = new Set([
  '/api/v1/auth/login',
  '/api/v1/auth/initialize',
])

/** 登录失败、初始化失败不应被当成会话失效。 */
export function isSessionNeutralRequest(url: string | undefined): boolean {
  if (!url) return false
  const path = url.split('?')[0] ?? url
  return SESSION_NEUTRAL_URLS.has(path)
}

/**
 * 刷新失败后是否丢掉当前 Cookie。
 * 失败的是旧刷新令牌、而浏览器里已经是另一把时，保留新会话。
 */
export function shouldDiscardSessionAfterRefreshFailure(
  attemptedRefreshToken: string | undefined,
  currentRefreshToken: string | undefined,
): boolean {
  if (!currentRefreshToken) return true
  if (!attemptedRefreshToken) return false
  return attemptedRefreshToken === currentRefreshToken
}

/** 从刷新请求体取出这次真正提交的 refresh_token。 */
export function refreshTokenFromRequest(
  config:
    | {
        data?: unknown
      }
    | null
    | undefined,
): string | undefined {
  const data = config?.data
  if (typeof data === 'string') {
    try {
      return refreshTokenFromRequest({ data: JSON.parse(data) as unknown })
    } catch {
      return undefined
    }
  }
  if (!data || typeof data !== 'object') return undefined
  const token = (data as { refresh_token?: unknown }).refresh_token
  return typeof token === 'string' && token !== '' ? token : undefined
}

/**
 * 刷新令牌还在、且访问令牌即将到期或到期时间已不可读时，需要续期。
 */
export function shouldRefreshAccessToken(
  expiresAtRaw: string | undefined,
  nowMs: number,
  hasRefreshToken: boolean,
): boolean {
  if (!hasRefreshToken) return false
  if (!expiresAtRaw) return true
  const expiresAtMs = Number(expiresAtRaw)
  if (Number.isNaN(expiresAtMs)) return true
  return expiresAtMs - nowMs < REFRESH_LEAD_MS
}
