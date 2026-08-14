import Cookies from 'js-cookie'

export const THIRTY_DAYS_IN_SECONDS = 30 * 24 * 60 * 60

// HTTPS 部署下标记 Cookie 为 Secure，防止明文 HTTP 请求携带凭据被中间人截获
const isSecure = typeof window !== 'undefined' && window.location.protocol === 'https:'

export function writeTokenCookies(
  tokenData: {
    access_token: string
    refresh_token: string
    expires_in: number
  },
  options?: {
    accessTokenExpiresInDays?: number
  },
) {
  const accessTokenExpires =
    options?.accessTokenExpiresInDays ?? tokenData.expires_in / 86400

  Cookies.set('access_token', tokenData.access_token, {
    expires: accessTokenExpires,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
  Cookies.set('refresh_token', tokenData.refresh_token, {
    expires: THIRTY_DAYS_IN_SECONDS / 86400,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
  Cookies.set('expires_at', String(Date.now() + tokenData.expires_in * 1000), {
    expires: tokenData.expires_in / 86400,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
}
