/** @vitest-environment jsdom */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Cookies from 'js-cookie'
import { writeTokenCookies } from './cookie-utils'

describe('writeTokenCookies', () => {
  beforeEach(() => {
    const all = Cookies.get()
    for (const key of Object.keys(all)) {
      Cookies.remove(key, { path: '/' })
    }
    vi.restoreAllMocks()
  })

  it('Q-04/Q-05: access cookie follows expires_in, refresh and expires_at follow refresh_expires_in', () => {
    const setSpy = vi.spyOn(Cookies, 'set')

    writeTokenCookies({
      access_token: 'at-1',
      refresh_token: 'rt-1',
      expires_in: 3600,
      refresh_expires_in: 604800,
    })

    expect(setSpy).toHaveBeenCalledWith(
      'access_token',
      'at-1',
      expect.objectContaining({ expires: 3600 / 86400 }),
    )
    expect(setSpy).toHaveBeenCalledWith(
      'refresh_token',
      'rt-1',
      expect.objectContaining({ expires: 604800 / 86400 }),
    )
    expect(setSpy).toHaveBeenCalledWith(
      'expires_at',
      expect.any(String),
      expect.objectContaining({ expires: 604800 / 86400 }),
    )
  })
})
