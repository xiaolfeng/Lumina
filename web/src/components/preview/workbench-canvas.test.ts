import { describe, expect, it } from 'vitest'

import { clampDeviceWidth } from './workbench-canvas'

describe('clampDeviceWidth', () => {
  it('keeps the fixed 320 lower bound', () => {
    expect(clampDeviceWidth(100, 1440)).toBe(320)
    expect(clampDeviceWidth(-5, 1440)).toBe(320)
  })

  it('caps width at min(1280, viewport - 32) on wide viewports', () => {
    expect(clampDeviceWidth(2000, 1920)).toBe(1280)
    expect(clampDeviceWidth(1200, 1024)).toBe(992)
    expect(clampDeviceWidth(375, 1024)).toBe(375)
  })

  it('never lets the dynamic upper bound fall below the lower bound', () => {
    expect(clampDeviceWidth(1200, 300)).toBe(320)
  })

  it('rounds fractional drag positions to integers', () => {
    expect(clampDeviceWidth(500.6, 800)).toBe(501)
  })
})
