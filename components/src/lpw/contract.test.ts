import { describe, expect, it } from 'vitest'
import {
  blockTypeGroups,
  containerVariants,
  layoutPatterns,
} from './contract'

describe('contract', () => {
  it('container variants snapshot matches Go contract', () => {
    const sorted = Object.fromEntries(
      Object.entries(containerVariants)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([k, v]) => [
          k,
          Object.fromEntries(
            Object.entries(v).sort(([a], [b]) => a.localeCompare(b)),
          ),
        ]),
    )
    expect(sorted).toMatchSnapshot()
  })

  it('block groups snapshot matches Go contract', () => {
    const sorted = Object.fromEntries(
      Object.entries(blockTypeGroups).sort(([a], [b]) => a.localeCompare(b)),
    )
    expect(sorted).toMatchSnapshot()
    expect(blockTypeGroups.glance).toEqual(['decision'])
    expect(blockTypeGroups.personnel).toEqual(['process', 'text'])
  })

  it('layout patterns snapshot matches Go contract', () => {
    const sorted = Object.fromEntries(
      Object.entries(layoutPatterns).sort(([a], [b]) => a.localeCompare(b)),
    )
    expect(sorted).toMatchSnapshot()
  })
})
