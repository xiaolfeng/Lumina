import { describe, expect, it } from 'vitest'
import { isLpwIconName, LPW_ICON_NAMES } from './icon-map'

describe('icon-map', () => {
  it('identifies valid icon names', () => {
    expect(isLpwIconName('bookmark')).toBe(true)
    expect(isLpwIconName('nope')).toBe(false)
    expect(isLpwIconName(42)).toBe(false)
    expect(isLpwIconName(null)).toBe(false)
  })

  it('has 10 controlled icons', () => {
    expect(LPW_ICON_NAMES.length).toBe(10)
  })
})
