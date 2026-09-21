import { describe, expect, it } from 'vitest'
import { parseLpwSource } from './parser'

describe('parser', () => {
  it('parses valid 1.1 doc', () => {
    const src = JSON.stringify({
      version: '1.1',
      meta: { title: '测试文档' },
      content: [
        {
          id: 'b1',
          kind: 'block',
          type: 'markdown',
          props: { content: 'hello' },
        },
      ],
    })
    const res = parseLpwSource(src)
    expect(res.error).toBeUndefined()
    expect(res.document?.version).toBe('1.1')
    expect(res.document?.content.length).toBe(1)
  })

  it('rejects version 1.0', () => {
    const src = JSON.stringify({
      version: '1.0',
      content: [],
    })
    const res = parseLpwSource(src)
    expect(res.document).toBeUndefined()
    expect(res.error?.code).toBe('UNSUPPORTED_VERSION')
    expect(res.error?.path).toBe('/version')
  })

  it('rejects missing kind with path', () => {
    const src = JSON.stringify({
      version: '1.1',
      content: [
        { id: 'b1', type: 'markdown', props: { content: 'x' } },
      ],
    })
    const res = parseLpwSource(src)
    expect(res.document).toBeUndefined()
    expect(res.error?.code).toBe('INVALID_STRUCTURE')
    expect(res.error?.path).toBe('/content/0/kind')
  })

  it('rejects duplicate ids at second occurrence', () => {
    const src = JSON.stringify({
      version: '1.1',
      content: [
        { id: 'dup', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'dup', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    })
    const res = parseLpwSource(src)
    expect(res.document).toBeUndefined()
    expect(res.error?.code).toBe('DUPLICATE_ID')
    expect(res.error?.path).toBe('/content/1/id')
  })

  it('unknown type is not a parse error', () => {
    const src = JSON.stringify({
      version: '1.1',
      content: [
        { id: 'b1', kind: 'block', type: 'future-block', props: {} },
      ],
    })
    const res = parseLpwSource(src)
    expect(res.error).toBeUndefined()
    expect(res.document?.content[0].type).toBe('future-block')
  })
})
