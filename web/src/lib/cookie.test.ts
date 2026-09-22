/** @vitest-environment jsdom */
import { describe, expect, it, beforeEach, vi } from 'vitest'
import Cookies from 'js-cookie'
import {
  COOKIE_KEYS,
  getCookie,
  setCookie,
  removeCookie,
  getQaSplitterRatio,
  setQaSplitterRatio,
  QA_SPLITTER_COOKIE_EXPIRES_DAYS,
  QA_SPLITTER_DEFAULT_RATIO,
  QA_SPLITTER_MIN_RATIO,
  QA_SPLITTER_MAX_RATIO,
} from './cookie'

describe('cookie governance', () => {
  beforeEach(() => {
    // 清理所有 cookie
    const all = Cookies.get()
    for (const key of Object.keys(all)) {
      Cookies.remove(key, { path: '/' })
    }
  })

  it('getCookie / setCookie / removeCookie handles basic cookie lifecycle', () => {
    expect(getCookie(COOKIE_KEYS.ACCESS_TOKEN)).toBeUndefined()

    setCookie(COOKIE_KEYS.ACCESS_TOKEN, 'test_token_val')
    expect(getCookie(COOKIE_KEYS.ACCESS_TOKEN)).toBe('test_token_val')

    removeCookie(COOKIE_KEYS.ACCESS_TOKEN)
    expect(getCookie(COOKIE_KEYS.ACCESS_TOKEN)).toBeUndefined()
  })

  it('setCookie passes default options including path and sameSite', () => {
    const setSpy = vi.spyOn(Cookies, 'set')
    setCookie('custom_key', 'custom_val', { expires: 7 })

    expect(setSpy).toHaveBeenCalledWith(
      'custom_key',
      'custom_val',
      expect.objectContaining({
        path: '/',
        sameSite: 'Lax',
        expires: 7,
      }),
    )
    setSpy.mockRestore()
  })

  describe('Q&A Splitter Ratio Cookie', () => {
    it('returns null when cookie is absent', () => {
      expect(QA_SPLITTER_DEFAULT_RATIO).toBe(45)
      expect(getQaSplitterRatio()).toBeNull()
    })

    it('saves ratio with clamped range and expires days', () => {
      const setSpy = vi.spyOn(Cookies, 'set')

      // 正常合法值 38%
      setQaSplitterRatio(38)
      expect(setSpy).toHaveBeenCalledWith(
        COOKIE_KEYS.QA_SPLITTER_RATIO,
        '38',
        expect.objectContaining({
          expires: QA_SPLITTER_COOKIE_EXPIRES_DAYS,
          path: '/',
          sameSite: 'Lax',
        }),
      )
      expect(getQaSplitterRatio()).toBe(38)

      // 小于最小值 30 时 clamped 到 30
      setQaSplitterRatio(15)
      expect(getQaSplitterRatio()).toBe(QA_SPLITTER_MIN_RATIO)

      // 大于最大值 50 时 clamped 到 50
      setQaSplitterRatio(75)
      expect(getQaSplitterRatio()).toBe(QA_SPLITTER_MAX_RATIO)

      // 浮点数四舍五入
      setQaSplitterRatio(42.6)
      expect(getQaSplitterRatio()).toBe(43)

      setSpy.mockRestore()
    })

    it('returns null if cookie value is invalid or out of range', () => {
      // 恶意或异常非数值
      Cookies.set(COOKIE_KEYS.QA_SPLITTER_RATIO, 'invalid_number', { path: '/' })
      expect(getQaSplitterRatio()).toBeNull()

      // 外部篡改为超出界限的值
      Cookies.set(COOKIE_KEYS.QA_SPLITTER_RATIO, '10', { path: '/' })
      expect(getQaSplitterRatio()).toBeNull()

      Cookies.set(COOKIE_KEYS.QA_SPLITTER_RATIO, '90', { path: '/' })
      expect(getQaSplitterRatio()).toBeNull()
    })
  })
})
