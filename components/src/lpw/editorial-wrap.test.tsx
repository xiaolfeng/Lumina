/** @vitest-environment jsdom */
import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { renderLayoutPreset } from './layout/presets'
import type { LpwLayoutChild, LpwLayoutProps } from './types'

afterEach(() => {
  cleanup()
})

describe('editorial-wrap', () => {
  const rawChildren: LpwLayoutChild[] = [
    {
      id: 'img-1',
      kind: 'block',
      type: 'image',
      props: { src: 'cover.png', alt: '封面图' },
    },
    {
      id: 'md-1',
      kind: 'block',
      type: 'markdown',
      props: { content: '正文说明段落' },
    },
  ]

  const renderedChildren = [
    <div key="img" data-testid="img-node">
      [封面图]
    </div>,
    <div key="md" data-testid="md-node">
      正文说明段落
    </div>,
  ]

  it('desktop uses float strategy', () => {
    const props: LpwLayoutProps = {
      pattern: 'editorial-wrap',
      strategy: {
        type: 'media-wrap',
        mediaPosition: 'top-start',
        mediaWidth: 'third',
        mediaShape: 'square',
      },
    }

    const { container } = render(
      <div>
        {renderLayoutPreset({
          props,
          rawChildren,
          renderedChildren,
        })}
      </div>,
    )

    const wrap = container.querySelector('[data-testid="layout-editorial-wrap"]')
    expect(wrap).toBeTruthy()

    // 桌面端浮动包裹层（单 DOM 树，配合 float 与 CSS 变量）
    const floatDiv = wrap?.firstElementChild as HTMLElement
    expect(floatDiv).toBeTruthy()
    expect(floatDiv.style.float).toBe('left')
    expect(floatDiv.style.getPropertyValue('--media-width')).toBe('33.333%')
  })

  it('mobile stacks by orderOnMobile', () => {
    // 设置 img 顺序为 2，md 顺序为 1
    const props: LpwLayoutProps = {
      pattern: 'editorial-wrap',
      placements: [
        { nodeId: 'img-1', orderOnMobile: 2 },
        { nodeId: 'md-1', orderOnMobile: 1 },
      ],
    }

    const { container } = render(
      <div>
        {renderLayoutPreset({
          props,
          rawChildren,
          renderedChildren,
        })}
      </div>,
    )

    const wrap = container.querySelector(
      '[data-testid="layout-editorial-wrap"]',
    ) as HTMLElement
    expect(wrap).toBeTruthy()
    expect(wrap.className).toContain('flex')
    expect(wrap.className).toContain('flex-col')
    expect(wrap.className).toContain('@3xl/lpw:block')

    const imgWrap = wrap.children[0] as HTMLElement
    const mdWrap = wrap.children[1] as HTMLElement
    expect(imgWrap.className).toContain('@max-3xl/lpw:!mx-0')
    expect(imgWrap.className).toContain('@max-3xl/lpw:!w-full')
    expect(imgWrap.style.getPropertyValue('--m-order')).toBe('2')
    expect(mdWrap.style.getPropertyValue('--m-order')).toBe('1')
  })

  it('Q-06: 消除桌面端与移动端双重 JSX 挂载，收敛为单节点 DOM 树', () => {
    const props: LpwLayoutProps = {
      pattern: 'editorial-wrap',
      placements: [
        { nodeId: 'img-1', orderOnMobile: 1 },
        { nodeId: 'md-1', orderOnMobile: 2 },
      ],
    }

    const { container } = render(
      <div>
        {renderLayoutPreset({
          props,
          rawChildren,
          renderedChildren,
        })}
      </div>,
    )

    // 在整个 editorial-wrap 内，每个子元素（img-node, md-node）应当仅在 DOM 树中出现一次，杜绝双重挂载造成 DOM ID 冲突
    const imgNodes = container.querySelectorAll('[data-testid="img-node"]')
    const mdNodes = container.querySelectorAll('[data-testid="md-node"]')
    expect(imgNodes.length).toBe(1)
    expect(mdNodes.length).toBe(1)
  })
})
