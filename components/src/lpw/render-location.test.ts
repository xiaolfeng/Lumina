import { describe, expect, it } from 'vitest'
import {
  childLocation,
  formatLocation,
  rootLocation,
} from './render-location'

describe('render-location', () => {
  it('rootLocation returns /content', () => {
    const root = rootLocation()
    expect(root.jsonPath).toBe('/content')
    expect(root.idPath).toEqual([])
    expect(formatLocation(root)).toBe('/content')
  })

  it('childLocation appends index and id', () => {
    const root = rootLocation()
    const child = childLocation(root, 0, 'node-1')
    expect(child.jsonPath).toBe('/content/0')
    expect(child.idPath).toEqual(['node-1'])
    expect(formatLocation(child)).toBe('/content/0（node-1）')

    const grandChild = childLocation(child, 2, 'child-2')
    expect(grandChild.jsonPath).toBe('/content/0/2')
    expect(grandChild.idPath).toEqual(['node-1', 'child-2'])
    expect(formatLocation(grandChild)).toBe('/content/0/2（node-1 › child-2）')
  })
})
