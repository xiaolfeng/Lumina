import { describe, expect, it } from 'vitest'
import { lpwRegistry, registerAll } from './index'

describe('LPW registerAll 对齐测试', () => {
  it('应准确注册全部 30 种块，且容器集合恰好是 4 种', () => {
    lpwRegistry.clear()
    expect(lpwRegistry.types()).toHaveLength(0)

    registerAll()
    const types = lpwRegistry.types()
    expect(types).toHaveLength(30)

    const expectedTypes = [
      // Plan 2 (9 种)
      'callout',
      'cards',
      'code',
      'divider',
      'heading',
      'image',
      'list',
      'markdown',
      'quote',
      // Plan 3 (10 种)
      'chart',
      'comparison',
      'diff',
      'mermaid',
      'metrics',
      'progress',
      'steps',
      'table',
      'timeline',
      'tree',
      // Plan 4 (7 种)
      'gallery',
      'glance',
      'open-items',
      'personnel',
      'quadrant',
      'scorecard',
      'takeaway',
      // Plan 5 (4 种容器)
      'columns',
      'details',
      'section',
      'tabs',
    ].sort()

    expect(types.sort()).toEqual(expectedTypes)

    // 容器判定断言
    const containerTypes = types
      .filter((t) => lpwRegistry.get(t)?.container === true)
      .sort()
    expect(containerTypes).toEqual(
      ['columns', 'details', 'section', 'tabs'].sort(),
    )

    // 非容器判定断言
    const leafTypes = types
      .filter((t) => lpwRegistry.get(t)?.container === false)
      .sort()
    expect(leafTypes).toHaveLength(26)
  })
})
