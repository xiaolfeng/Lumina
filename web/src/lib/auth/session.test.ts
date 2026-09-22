/** @vitest-environment jsdom */
import { describe, expect, it } from 'vitest'
import {
  isSessionNeutralRequest,
  refreshTokenFromRequest,
  shouldDiscardSessionAfterRefreshFailure,
  shouldRefreshAccessToken,
} from './session'

describe('auth session decisions', () => {
  it('Q-01: a failed refresh must not discard a token another refresh already installed', () => {
    expect(shouldDiscardSessionAfterRefreshFailure('rt-old', 'rt-new')).toBe(
      false,
    )
    expect(shouldDiscardSessionAfterRefreshFailure('rt-old', 'rt-old')).toBe(
      true,
    )
    expect(shouldDiscardSessionAfterRefreshFailure('rt-old', undefined)).toBe(
      true,
    )
    expect(shouldDiscardSessionAfterRefreshFailure(undefined, 'rt-new')).toBe(
      false,
    )
  })

  it('Q-01: login and initialize 401s are not session failures', () => {
    expect(isSessionNeutralRequest('/api/v1/auth/login')).toBe(true)
    expect(isSessionNeutralRequest('/api/v1/auth/initialize')).toBe(true)
    expect(isSessionNeutralRequest('/api/v1/auth/refresh')).toBe(false)
    expect(isSessionNeutralRequest('/api/v1/workspace')).toBe(false)
  })

  it('Q-01: reads the refresh token that this request actually sent', () => {
    expect(
      refreshTokenFromRequest({ data: { refresh_token: 'rt-body' } }),
    ).toBe('rt-body')
    expect(
      refreshTokenFromRequest({
        data: JSON.stringify({ refresh_token: 'rt-json' }),
      }),
    ).toBe('rt-json')
    expect(refreshTokenFromRequest({ data: '{}' })).toBeUndefined()
  })

  it('Q-03: missing or due expires_at still refreshes while a refresh token exists', () => {
    const now = 1_700_000_000_000
    expect(shouldRefreshAccessToken(undefined, now, true)).toBe(true)
    expect(shouldRefreshAccessToken('not-a-number', now, true)).toBe(true)
    expect(shouldRefreshAccessToken(String(now + 60_000), now, true)).toBe(true)
    expect(shouldRefreshAccessToken(String(now + 10 * 60_000), now, true)).toBe(
      false,
    )
    expect(shouldRefreshAccessToken(undefined, now, false)).toBe(false)
  })
})
