import { describe, expect, it } from 'vitest'
import { createMemoryStorage } from './test-setup'

describe('createMemoryStorage（test-setup 兜底实现）', () => {
  it('完整实现 Storage 接口语义', () => {
    const s = createMemoryStorage()

    expect(s.length).toBe(0)
    expect(s.getItem('missing')).toBeNull()
    expect(s.key(0)).toBeNull()

    s.setItem('a', '1')
    s.setItem('b', 2 as unknown as string) // 非字符串值被 String 强转
    expect(s.length).toBe(2)
    expect(s.getItem('a')).toBe('1')
    expect(s.getItem('b')).toBe('2')
    expect(s.key(0)).toBe('a')
    expect(s.key(1)).toBe('b')
    expect(s.key(9)).toBeNull()

    s.removeItem('a')
    expect(s.getItem('a')).toBeNull()
    expect(s.length).toBe(1)

    s.clear()
    expect(s.length).toBe(0)
    expect(s.getItem('b')).toBeNull()
  })

  it('每次调用返回独立实例（无跨实例共享状态）', () => {
    const a = createMemoryStorage()
    const b = createMemoryStorage()
    a.setItem('k', 'v')
    expect(b.getItem('k')).toBeNull()
  })
})
