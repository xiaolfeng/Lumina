/** @vitest-environment jsdom */
import { cleanup, render } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it } from 'vitest'
import type { LpwLayoutChild, LpwLayoutProps } from '../types'
import { renderLayoutPreset } from './presets'

afterEach(() => {
  cleanup()
})

function pair(
  ids: [string, string],
): { raw: LpwLayoutChild[]; rendered: ReactNode[] } {
  const raw: LpwLayoutChild[] = ids.map((id) => ({
    id,
    kind: 'block',
    type: 'markdown',
    props: { content: id },
  }))
  const rendered = ids.map((id) => (
    <div key={id} data-testid={`child-${id}`}>
      {id}
    </div>
  ))
  return { raw, rendered }
}

describe('Q-01 split / bento 窄屏列模板', () => {
  it('split 不在元素 style 上写死两列，窄屏 class 保持单列', () => {
    const { raw, rendered } = pair(['a', 'b'])
    const props: LpwLayoutProps = { pattern: 'split' }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const el = container.querySelector('[data-testid="layout-split"]') as HTMLElement
    expect(el).toBeTruthy()
    expect(el.className).toContain('grid-cols-1')
    expect(el.style.gridTemplateColumns).toBe('')
    expect(el.style.getPropertyValue('--layout-cols')).toBe(
      'minmax(0, 1fr) minmax(0, 1fr)',
    )
  })

  it('split fixed-fluid 把 360px 放进 CSS 变量而不是直接覆盖 grid-template-columns', () => {
    const { raw, rendered } = pair(['a', 'b'])
    const props: LpwLayoutProps = {
      pattern: 'split',
      strategy: { type: 'fixed-fluid', fixed: 'end', size: 'md' },
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const el = container.querySelector('[data-testid="layout-split"]') as HTMLElement
    expect(el.style.gridTemplateColumns).toBe('')
    expect(el.style.getPropertyValue('--layout-cols')).toBe('minmax(0, 1fr) 360px')
    expect(el.className).toContain('grid-cols-1')
  })

  it('bento 窄屏不写死 colSpan / 多列模板', () => {
    const { raw, rendered } = pair(['a', 'b'])
    const props: LpwLayoutProps = {
      pattern: 'bento',
      placements: [{ nodeId: 'a', colSpan: 2, rowSpan: 1 }],
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const el = container.querySelector('[data-testid="layout-bento"]') as HTMLElement
    expect(el.style.gridTemplateColumns).toBe('')
    expect(el.className).toContain('grid-cols-1')
    const firstCell = el.firstElementChild as HTMLElement
    expect(firstCell.style.gridColumn).toBe('')
    expect(firstCell.className).toContain('md:col-span-2')
  })
})

describe('Q-02 newspaper 正文可跨栏', () => {
  it('break-inside-avoid 只包 aside，不包 body', () => {
    const raw: LpwLayoutChild[] = [
      { id: 'body', kind: 'block', type: 'markdown', props: { content: 'body' } },
      { id: 'aside', kind: 'block', type: 'quote', props: { content: 'aside' } },
    ]
    const rendered = [
      <div key="body">body</div>,
      <div key="aside">aside</div>,
    ]
    const props: LpwLayoutProps = {
      pattern: 'newspaper',
      placements: [
        { nodeId: 'body', role: 'body' },
        { nodeId: 'aside', role: 'aside' },
      ],
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const body = container.querySelector(
      '[data-testid="layout-newspaper-body"]',
    ) as HTMLElement
    const aside = container.querySelector(
      '[data-testid="layout-newspaper-aside"]',
    ) as HTMLElement
    expect(body).toBeTruthy()
    expect(aside).toBeTruthy()
    expect(body.className).not.toContain('break-inside-avoid')
    expect(aside.className).toContain('break-inside-avoid')
    const flow = container.querySelector(
      '[data-testid="layout-newspaper-flow"]',
    ) as HTMLElement
    expect(flow.className).not.toContain('[&>*]:break-inside-avoid')
  })

  it('strategy.columns count=2 不会升到三栏', () => {
    const raw: LpwLayoutChild[] = [
      { id: 'body', kind: 'block', type: 'markdown', props: { content: 'body' } },
      { id: 'aside', kind: 'block', type: 'quote', props: { content: 'aside' } },
    ]
    const rendered = [<div key="body">body</div>, <div key="aside">aside</div>]
    const props: LpwLayoutProps = {
      pattern: 'newspaper',
      strategy: { type: 'columns', count: 2 },
      placements: [
        { nodeId: 'body', role: 'body' },
        { nodeId: 'aside', role: 'aside' },
      ],
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const flow = container.querySelector(
      '[data-testid="layout-newspaper-flow"]',
    ) as HTMLElement
    expect(flow.className).toContain('md:columns-2')
    expect(flow.className).not.toContain('lg:columns-3')
  })
})

describe('Q-03 alternating 窄屏保持语义 DOM 顺序', () => {
  it('奇数镜像组不交换 DOM，只用 md:order 做桌面镜像', () => {
    const ids = ['m1', 'c1', 'm2', 'c2'] as const
    const raw: LpwLayoutChild[] = ids.map((id) => ({
      id,
      kind: 'block',
      type: 'markdown',
      props: { content: id },
    }))
    const rendered = ids.map((id) => (
      <div key={id} data-testid={`child-${id}`}>
        {id}
      </div>
    ))
    const props: LpwLayoutProps = {
      pattern: 'alternating',
      alternateFrom: 'media',
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const groups = container.querySelectorAll(
      '[data-testid="layout-alternating"] > div',
    )
    const odd = groups[1] as HTMLElement
    expect(odd.textContent).toBe('m2c2')
    const first = odd.children[0] as HTMLElement
    const second = odd.children[1] as HTMLElement
    expect(first.className).toContain('md:order-2')
    expect(second.className).toContain('md:order-1')
  })
})

describe('Q-08 orderOnMobile 写入窄屏 order', () => {
  it('split 子节点按 placements.orderOnMobile 设置 --m-order', () => {
    const { raw, rendered } = pair(['a', 'b'])
    const props: LpwLayoutProps = {
      pattern: 'split',
      placements: [
        { nodeId: 'a', orderOnMobile: 2 },
        { nodeId: 'b', orderOnMobile: 1 },
      ],
    }
    const { container } = render(
      <>{renderLayoutPreset({ props, rawChildren: raw, renderedChildren: rendered })}</>,
    )
    const cells = container.querySelectorAll(
      '[data-testid="layout-split"] > div',
    )
    expect((cells[0] as HTMLElement).style.getPropertyValue('--m-order')).toBe(
      '2',
    )
    expect((cells[1] as HTMLElement).style.getPropertyValue('--m-order')).toBe(
      '1',
    )
  })
})
