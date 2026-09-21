import { describe, expect, it } from 'vitest'
import { lpwRegistry, registerAll } from './index'

describe('LPW registerAll 对齐测试', () => {
  it('应准确注册全部 31 种组件（1 种 Layout + 4 种 Container + 26 种 Block），且无 columns', () => {
    lpwRegistry.clear()
    expect(lpwRegistry.entries()).toHaveLength(0)

    registerAll()
    const all = lpwRegistry.entries()
    expect(all).toHaveLength(31)

    // Layout
    const layouts = lpwRegistry.entries('layout')
    expect(layouts).toHaveLength(1)
    expect(layouts[0].displayName).toBe('Layout')

    // Containers
    const containers = lpwRegistry.entries('container')
    expect(containers).toHaveLength(4)
    const containerTypes = containers.map((c) => c.displayName)
    expect(containerTypes).toContain('Section')
    expect(containerTypes).toContain('Panel')
    expect(containerTypes).toContain('Details')
    expect(containerTypes).toContain('Tabs')
    expect(containerTypes).not.toContain('Columns')

    // Blocks
    const blocks = lpwRegistry.entries('block')
    expect(blocks).toHaveLength(26)

    // columns 不在任何类别中
    expect(lpwRegistry.get('container', 'columns')).toBeUndefined()
  })
})
