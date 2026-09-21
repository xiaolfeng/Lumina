import { describe, expect, it } from 'vitest'
import { isSafeAnnotationPattern } from './safe-pattern'

describe('isSafeAnnotationPattern', () => {
  it('rejects nested quantifiers (catastrophic backtracking)', () => {
    expect(isSafeAnnotationPattern('(a+)+$')).toBe(false)
    expect(isSafeAnnotationPattern('(a+)+b')).toBe(false)
    expect(isSafeAnnotationPattern('(\\d{2,3})+')).toBe(false)
    expect(isSafeAnnotationPattern('((a b)+)+')).toBe(false)
    expect(isSafeAnnotationPattern('(a*)*b')).toBe(false)
    expect(isSafeAnnotationPattern('(\\p{L}+)+')).toBe(false)
    expect(isSafeAnnotationPattern('(\\u{1F600}+)+')).toBe(false)
  })

  it('rejects quantified alternation groups', () => {
    expect(isSafeAnnotationPattern('(a|a)*')).toBe(false)
    expect(isSafeAnnotationPattern('(?:Redis|Memcached)+')).toBe(false)
    expect(isSafeAnnotationPattern('(v1|v2)+')).toBe(false)
  })

  it('allows common safe patterns', () => {
    expect(isSafeAnnotationPattern('重要结论')).toBe(true)
    expect(isSafeAnnotationPattern('(Redis)\\s(缓存)')).toBe(true)
    expect(isSafeAnnotationPattern('[a-z]+')).toBe(true)
    expect(isSafeAnnotationPattern('(?:ab)+c')).toBe(true)
    expect(isSafeAnnotationPattern('(?:jpg|png)')).toBe(true)
    expect(isSafeAnnotationPattern('^\\d{4}-\\d{2}-\\d{2}$')).toBe(true)
    expect(isSafeAnnotationPattern('(\\w+\\s*)')).toBe(true)
    expect(isSafeAnnotationPattern('[a+]+')).toBe(true) // 字符类内的 + 是字面量
    expect(isSafeAnnotationPattern('(\\p{L})+')).toBe(true) // Q-09: Unicode 属性转义整体消费为字符原子
    expect(isSafeAnnotationPattern('(\\u{1F600})+')).toBe(true) // Q-09: Unicode 码点转义整体消费为字符原子
  })

  it('rejects malformed or oversized patterns', () => {
    expect(isSafeAnnotationPattern('')).toBe(false)
    expect(isSafeAnnotationPattern('(')).toBe(false)
    expect(isSafeAnnotationPattern(')')).toBe(false)
    expect(isSafeAnnotationPattern('a'.repeat(129))).toBe(false)
    expect(isSafeAnnotationPattern('a'.repeat(128))).toBe(true)
  })

  it('treats lookaround groups as groups for nesting checks', () => {
    expect(isSafeAnnotationPattern('(?=(a+)+)b')).toBe(false)
    expect(isSafeAnnotationPattern('(?=\\d+)元')).toBe(true)
  })
})
