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

    // 桌面端浮动包裹层
    const desktopContainer = wrap?.querySelector('.hidden.md\\:block')
    expect(desktopContainer).toBeTruthy()
    const floatDiv = desktopContainer?.firstElementChild as HTMLElement
    expect(floatDiv.style.float).toBe('left')
    expect(floatDiv.style.width).toBe('33.333%')
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

    const mobileDiv = container.querySelector(
      '[data-testid="layout-editorial-wrap"] .md\\:hidden',
    )
    expect(mobileDiv).toBeTruthy()
    // 断言 DOM 文本顺序：正文在前，图片在后
    const text = mobileDiv?.textContent || ''
    expect(text.indexOf('正文说明段落')).toBeLessThan(text.indexOf('[封面图]'))
  })
})
