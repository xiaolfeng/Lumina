import Cookies from 'js-cookie'

/**
 * 集中管理 Cookie 键名常量，避免硬编码与拼写冲突
 */
export const COOKIE_KEYS = {
  // 认证与令牌
  ACCESS_TOKEN: 'access_token',
  REFRESH_TOKEN: 'refresh_token',
  EXPIRES_AT: 'expires_at',
  // UI 偏好设置
  QA_SPLITTER_RATIO: 'lumina_qa_splitter_ratio',
} as const

export type KnownCookieKey = (typeof COOKIE_KEYS)[keyof typeof COOKIE_KEYS]
export type CookieKey = KnownCookieKey | (string & {})

/**
 * 安全策略：在 HTTPS 部署下标记 Cookie 为 Secure，防止明文传输被中间人截获
 */
const isSecure =
  typeof window !== 'undefined' && window.location.protocol === 'https:'

/**
 * 默认 Cookie 配置
 */
export const DEFAULT_COOKIE_OPTIONS: Cookies.CookieAttributes = {
  path: '/',
  sameSite: 'Lax',
  secure: isSecure,
}

/**
 * 读取 Cookie 值
 */
export function getCookie(key: CookieKey): string | undefined {
  return Cookies.get(key)
}

/**
 * 设置 Cookie
 */
export function setCookie(
  key: CookieKey,
  value: string,
  options?: Cookies.CookieAttributes,
): void {
  Cookies.set(key, value, {
    ...DEFAULT_COOKIE_OPTIONS,
    ...options,
  })
}

/**
 * 删除 Cookie
 */
export function removeCookie(
  key: CookieKey,
  options?: Cookies.CookieAttributes,
): void {
  Cookies.remove(key, {
    ...DEFAULT_COOKIE_OPTIONS,
    ...options,
  })
}

/* ────────────────────────────────────────────────────────────
   Q&A Splitter 左右分栏占比 Cookie 治理
   ──────────────────────────────────────────────────────────── */

/** Q&A 布局偏好 Cookie 有效期：90 天 */
export const QA_SPLITTER_COOKIE_EXPIRES_DAYS = 90
/** Q&A 默认左侧占比（%） */
export const QA_SPLITTER_DEFAULT_RATIO = 45
/** Q&A 左侧最小占比（%） */
export const QA_SPLITTER_MIN_RATIO = 30
/** Q&A 左侧最大占比（%） */
export const QA_SPLITTER_MAX_RATIO = 50

/**
 * 读取已保存的 Q&A 左右分栏左侧占比（%）
 * 若 Cookie 不存在、解析失败或超出合法 [30, 50] 范围，返回 null 由调用方回退默认值
 */
export function getQaSplitterRatio(): number | null {
  const raw = getCookie(COOKIE_KEYS.QA_SPLITTER_RATIO)
  if (!raw) return null

  const num = Number.parseFloat(raw)
  if (
    Number.isNaN(num) ||
    num < QA_SPLITTER_MIN_RATIO ||
    num > QA_SPLITTER_MAX_RATIO
  ) {
    return null
  }

  return Math.round(num)
}

/**
 * 保存 Q&A 左右分栏左侧占比（%）到 Cookie
 * 强制 clamp 到 [30, 50] 之间，并四舍五入为整型百分比
 */
export function setQaSplitterRatio(ratio: number): void {
  const clamped = Math.min(
    QA_SPLITTER_MAX_RATIO,
    Math.max(QA_SPLITTER_MIN_RATIO, Math.round(ratio)),
  )

  setCookie(COOKIE_KEYS.QA_SPLITTER_RATIO, String(clamped), {
    expires: QA_SPLITTER_COOKIE_EXPIRES_DAYS,
  })
}
