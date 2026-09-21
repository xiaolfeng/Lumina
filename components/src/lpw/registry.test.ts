import { beforeEach, describe, expect, it } from 'vitest'
import { lpwRegistry, registerAll } from './index'

describe('registry', () => {
  beforeEach(() => {
    lpwRegistry.clear()
    registerAll()
  })

  it('registers 31 entries', () => {
    const all = lpwRegistry.entries()
    expect(all.length).toBe(31)

    const layouts = lpwRegistry.entries('layout')
    expect(layouts.length).toBe(1)

    const containers = lpwRegistry.entries('container')
    expect(containers.length).toBe(4)

    const blocks = lpwRegistry.entries('block')
    expect(blocks.length).toBe(26)
  })

  it('columns absent in any kind', () => {
    expect(lpwRegistry.get('container', 'columns')).toBeUndefined()
    expect(lpwRegistry.get('layout', 'columns')).toBeUndefined()
    expect(lpwRegistry.get('block', 'columns')).toBeUndefined()
  })

  it('kind:type key prevents cross-kind collision', () => {
    // 假定注册一个同名 block 和 container
    lpwRegistry.register('block', 'collision-test', {
      displayName: 'Block Collision',
      Component: (() => null) as never,
      groups: ['text'],
      annotatableFields: [],
    })
    lpwRegistry.register('container', 'collision-test', {
      displayName: 'Container Collision',
      Component: (() => null) as never,
      variants: {},
    })

    const b = lpwRegistry.get('block', 'collision-test')
    const c = lpwRegistry.get('container', 'collision-test')

    expect(b?.kind).toBe('block')
    expect(c?.kind).toBe('container')
    expect(b?.displayName).toBe('Block Collision')
    expect(c?.displayName).toBe('Container Collision')
  })
})
