import { beforeEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from './lpw-registry'

describe('lpwRegistry', () => {
  beforeEach(() => {
    lpwRegistry.clear()
  })

  it('应能正确注册并获取组件', () => {
    const DummyComponent = () => null
    lpwRegistry.register('test-block', false, DummyComponent)

    const entry = lpwRegistry.get('test-block')
    expect(entry).toBeDefined()
    expect(entry?.container).toBe(false)
    expect(entry?.Component).toBe(DummyComponent)
  })

  it('获取未注册组件应返回 undefined', () => {
    expect(lpwRegistry.get('non-exist')).toBeUndefined()
  })

  it('覆盖注册时应更新条目', () => {
    const Comp1 = () => null
    const Comp2 = () => null
    lpwRegistry.register('item', false, Comp1)
    expect(lpwRegistry.get('item')?.Component).toBe(Comp1)

    lpwRegistry.register('item', true, Comp2)
    const entry = lpwRegistry.get('item')
    expect(entry?.container).toBe(true)
    expect(entry?.Component).toBe(Comp2)
  })

  it('types() 应返回已注册的所有类型名称', () => {
    expect(lpwRegistry.types()).toEqual([])
    lpwRegistry.register('markdown', false, () => null)
    lpwRegistry.register('section', true, () => null)
    expect(lpwRegistry.types()).toEqual(['markdown', 'section'])
  })
})
