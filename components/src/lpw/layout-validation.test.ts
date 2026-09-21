import { describe, expect, it } from 'vitest'
import { validateLayoutPattern } from './layout/validation'
import { rootLocation } from './render-location'
import type { LpwLayout } from './types'

describe('layout-validation', () => {
  const loc = rootLocation()

  it('split requires exactly 2', () => {
    const layout1: LpwLayout = {
      id: 'lay-1',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'split' },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
      ],
    }
    const d1 = validateLayoutPattern(layout1, loc)
    expect(d1.some((d) => d.reason.includes('超出限制'))).toBe(true)

    const layout2: LpwLayout = {
      id: 'lay-2',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'split' },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'b2', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    }
    const d2 = validateLayoutPattern(layout2, loc)
    expect(d2.length).toBe(0)
  })

  it('alternating rejects odd', () => {
    const layout: LpwLayout = {
      id: 'lay-alt',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'alternating' },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'b2', kind: 'block', type: 'markdown', props: { content: '2' } },
        { id: 'b3', kind: 'block', type: 'markdown', props: { content: '3' } },
      ],
    }
    const d = validateLayoutPattern(layout, loc)
    expect(d.some((x) => x.reason.includes('必须为偶数'))).toBe(true)
  })

  it('editorial-wrap only image plus markdown', () => {
    const layoutBad: LpwLayout = {
      id: 'lay-wrap',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'editorial-wrap' },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'b2', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    }
    const d = validateLayoutPattern(layoutBad, loc)
    expect(
      d.some((x) => x.reason.includes('必须恰好由一个 image 块和一个 markdown 块组成')),
    ).toBe(true)

    const layoutGood: LpwLayout = {
      id: 'lay-wrap-ok',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'editorial-wrap' },
      children: [
        { id: 'img', kind: 'block', type: 'image', props: { src: 'a.png' } },
        { id: 'md', kind: 'block', type: 'markdown', props: { content: 'txt' } },
      ],
    }
    expect(validateLayoutPattern(layoutGood, loc).length).toBe(0)
  })

  it('newspaper body must be markdown', () => {
    const layoutBad: LpwLayout = {
      id: 'lay-news',
      kind: 'layout',
      type: 'layout',
      props: {
        pattern: 'newspaper',
        placements: [
          { nodeId: 'ch-1', role: 'body' },
          { nodeId: 'md-2', role: 'full' },
        ],
      },
      children: [
        { id: 'ch-1', kind: 'block', type: 'chart', props: {} },
        { id: 'md-2', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    }
    const d = validateLayoutPattern(layoutBad, loc)
    expect(d.some((x) => x.reason.includes('role=body 的节点必须是 markdown 块'))).toBe(true)
  })

  it('newspaper full chart ok', () => {
    const layoutGood: LpwLayout = {
      id: 'lay-news-ok',
      kind: 'layout',
      type: 'layout',
      props: {
        pattern: 'newspaper',
        placements: [
          { nodeId: 'md-1', role: 'body' },
          { nodeId: 'ch-2', role: 'full' },
        ],
      },
      children: [
        { id: 'md-1', kind: 'block', type: 'markdown', props: { content: 'body' } },
        { id: 'ch-2', kind: 'block', type: 'chart', props: {} },
      ],
    }
    expect(validateLayoutPattern(layoutGood, loc).length).toBe(0)
  })

  it('placements foreign id rejected', () => {
    const layout: LpwLayout = {
      id: 'lay-p',
      kind: 'layout',
      type: 'layout',
      props: {
        pattern: 'split',
        placements: [{ nodeId: 'non-existing', role: 'primary' }],
      },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'b2', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    }
    const d = validateLayoutPattern(layout, loc)
    expect(d.some((x) => x.reason.includes('不是该 layout 的直接子节点'))).toBe(true)
  })

  it('orderOnMobile must be unique', () => {
    const layout: LpwLayout = {
      id: 'lay-order',
      kind: 'layout',
      type: 'layout',
      props: {
        pattern: 'split',
        placements: [
          { nodeId: 'b1', orderOnMobile: 1 },
          { nodeId: 'b2', orderOnMobile: 1 },
        ],
      },
      children: [
        { id: 'b1', kind: 'block', type: 'markdown', props: { content: '1' } },
        { id: 'b2', kind: 'block', type: 'markdown', props: { content: '2' } },
      ],
    }
    const d = validateLayoutPattern(layout, loc)
    expect(d.some((x) => x.reason.includes('orderOnMobile 1 重复'))).toBe(true)
  })
})
