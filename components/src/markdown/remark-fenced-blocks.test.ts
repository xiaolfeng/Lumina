import { describe, expect, it } from 'vitest'
import { remarkFencedBlocks } from './remark-fenced-blocks'

function createAst(children: unknown[]) {
  return { type: 'root', children }
}

function paragraph(text: string) {
  return { type: 'paragraph', children: [{ type: 'text', value: text }] }
}

describe('remarkFencedBlocks', () => {
  it('converts :::callout to custom callout node', () => {
    const tree = createAst([
      paragraph(':::callout'),
      paragraph('Hello world'),
      paragraph(':::'),
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'callout')
    expect(tree.children[0]).toHaveProperty('data.hName', 'callout')
    expect(tree.children[0]).toHaveProperty('data.hProperties', {})
    expect((tree.children[0] as { children: unknown[] }).children).toHaveLength(1)
    expect((tree.children[0] as { children: { type: string }[] }).children[0].type).toBe('paragraph')
  })

  it('converts :::callout with props to custom node with hProperties', () => {
    const tree = createAst([
      paragraph(':::callout{type="info"}'),
      paragraph('Hello world'),
      paragraph(':::'),
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'callout')
    expect(tree.children[0]).toHaveProperty('data.hProperties', { type: 'info' })
  })

  it('converts nested :::steps with :::step blocks', () => {
    const tree = createAst([
      paragraph(':::steps'),
      paragraph(':::step'),
      paragraph('Step 1'),
      paragraph(':::'),
      paragraph(':::step'),
      paragraph('Step 2'),
      paragraph(':::'),
      paragraph(':::'),
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'steps')
    const stepsChildren = (tree.children[0] as { children: unknown[] }).children
    expect(stepsChildren).toHaveLength(2)
    expect((stepsChildren[0] as { type: string }).type).toBe('step')
    expect((stepsChildren[1] as { type: string }).type).toBe('step')
    expect((stepsChildren[0] as { children: { type: string }[] }).children[0].type).toBe('paragraph')
    expect((stepsChildren[1] as { children: { type: string }[] }).children[0].type).toBe('paragraph')
  })

  it('treats unclosed fence as plain text', () => {
    const tree = createAst([
      paragraph(':::callout'),
      paragraph('Hello world'),
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(2)
    expect(tree.children[0]).toHaveProperty('type', 'paragraph')
    expect(tree.children[1]).toHaveProperty('type', 'paragraph')
  })

  it('does not affect code blocks', () => {
    const tree = createAst([
      { type: 'code', lang: 'markdown', value: ':::callout\nHello\n:::\n' },
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'code')
  })

  it('converts :::card with title and href props', () => {
    const tree = createAst([
      paragraph(':::card{title="My Card" href="https://example.com"}'),
      paragraph('Card content'),
      paragraph(':::'),
    ])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'card')
    expect(tree.children[0]).toHaveProperty('data.hProperties', {
      title: 'My Card',
      href: 'https://example.com',
    })
  })

  it('does not affect horizontal rules', () => {
    const tree = createAst([{ type: 'thematicBreak' }])

    const plugin = remarkFencedBlocks()
    plugin(tree)

    expect(tree.children).toHaveLength(1)
    expect(tree.children[0]).toHaveProperty('type', 'thematicBreak')
  })
})
