import { describe, expect, it } from 'vitest'
import { parseLpwSource } from './lpw-parser'

describe('parseLpwSource', () => {
  it('应成功解析合法文档（含嵌套 children）', () => {
    const valid = JSON.stringify({
      version: '1.0',
      meta: {
        title: '测试文档',
        tags: ['demo'],
      },
      blocks: [
        {
          id: 'intro',
          type: 'markdown',
          props: { content: 'hello' },
        },
        {
          id: 'sec-1',
          type: 'section',
          props: { title: '小节 1' },
          children: [
            {
              id: 'child-1',
              type: 'callout',
              props: { content: '提示内容' },
            },
          ],
        },
      ],
    })

    const result = parseLpwSource(valid)
    expect(result.error).toBeUndefined()
    expect(result.document).toBeDefined()
    expect(result.document?.version).toBe('1.0')
    expect(result.document?.blocks).toHaveLength(2)
    expect(result.document?.blocks[1].children?.[0].id).toBe('child-1')
  })

  it('非对象根节点应报错', () => {
    const arrayResult = parseLpwSource('[1, 2, 3]')
    expect(arrayResult.error).toBeDefined()
    expect(arrayResult.error?.message).toContain('根节点')

    const stringResult = parseLpwSource('"hello"')
    expect(stringResult.error).toBeDefined()
    expect(stringResult.error?.message).toContain('根节点')
  })

  it('错误版本应报错并包含不支持的 LPW 版本', () => {
    const res = parseLpwSource(JSON.stringify({ version: '2.0', blocks: [] }))
    expect(res.error).toBeDefined()
    expect(res.error?.path).toBe('version')
    expect(res.error?.message).toContain('不支持的 LPW 版本')
  })

  it('缺少 blocks 键或非数组应报错', () => {
    const missing = parseLpwSource(JSON.stringify({ version: '1.0' }))
    expect(missing.error).toBeDefined()
    expect(missing.error?.path).toBe('blocks')

    const notArray = parseLpwSource(
      JSON.stringify({ version: '1.0', blocks: 'invalid' }),
    )
    expect(notArray.error).toBeDefined()
    expect(notArray.error?.path).toBe('blocks')
  })

  it('块形状非法时应报错且 path 指向该块', () => {
    const missingId = parseLpwSource(
      JSON.stringify({
        version: '1.0',
        blocks: [{ type: 'markdown', props: {} }],
      }),
    )
    expect(missingId.error).toBeDefined()
    expect(missingId.error?.path).toBe('blocks[0].id')

    const typeNotString = parseLpwSource(
      JSON.stringify({
        version: '1.0',
        blocks: [{ id: 'b1', type: 123, props: {} }],
      }),
    )
    expect(typeNotString.error).toBeDefined()
    expect(typeNotString.error?.path).toBe('blocks[0].type')

    const propsNotObject = parseLpwSource(
      JSON.stringify({
        version: '1.0',
        blocks: [{ id: 'b1', type: 'markdown', props: 'not-object' }],
      }),
    )
    expect(propsNotObject.error).toBeDefined()
    expect(propsNotObject.error?.path).toBe('blocks[0].props')
  })

  it('id 重复（含 children）应定位到第二个出现处报错', () => {
    const dup = JSON.stringify({
      version: '1.0',
      blocks: [
        { id: 'same-id', type: 'markdown', props: {} },
        {
          id: 'sec-1',
          type: 'section',
          props: {},
          children: [{ id: 'same-id', type: 'callout', props: {} }],
        },
      ],
    })

    const res = parseLpwSource(dup)
    expect(res.error).toBeDefined()
    expect(res.error?.path).toBe('blocks[1].children[0]')
    expect(res.error?.message).toContain('重复')
  })

  it('JSON 语法错误应捕获并包含原始 message', () => {
    const res = parseLpwSource('{')
    expect(res.error).toBeDefined()
    expect(res.error?.message).toContain('JSON 语法解析失败')
  })

  it('未知类型不应报错（渲染期才走 Fallback）', () => {
    const res = parseLpwSource(
      JSON.stringify({
        version: '1.0',
        blocks: [{ id: 'b1', type: 'not-a-type', props: { foo: 'bar' } }],
      }),
    )
    expect(res.error).toBeUndefined()
    expect(res.document).toBeDefined()
    expect(res.document?.blocks[0].type).toBe('not-a-type')
  })
})
