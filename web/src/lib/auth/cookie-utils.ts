import Cookies from 'js-cookie'

/** 服务端未返回刷新令牌寿命时，与安全设置默认值 7 天对齐。 */
export const DEFAULT_REFRESH_COOKIE_SECONDS = 7 * 24 * 60 * 60

const isSecure =
  typeof window !== 'undefined' && window.location.protocol === 'https:'

export function writeTokenCookies(tokenData: {
  access_token: string
  refresh_token: string
  expires_in: number
  refresh_expires_in?: number
}) {
  const accessExpiresDays = tokenData.expires_in / 86400
  const refreshSeconds =
    tokenData.refresh_expires_in && tokenData.refresh_expires_in > 0
      ? tokenData.refresh_expires_in
      : DEFAULT_REFRESH_COOKIE_SECONDS
  const refreshExpiresDays = refreshSeconds / 86400

  Cookies.set('access_token', tokenData.access_token, {
    expires: accessExpiresDays,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
  Cookies.set('refresh_token', tokenData.refresh_token, {
    expires: refreshExpiresDays,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
  // 值是访问令牌到期时刻；Cookie 本身跟刷新令牌一样长，过期后仍能决定要不要续期。
  Cookies.set('expires_at', String(Date.now() + tokenData.expires_in * 1000), {
    expires: refreshExpiresDays,
    path: '/',
    sameSite: 'Lax',
    secure: isSecure,
  })
}
