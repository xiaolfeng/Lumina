/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { CalloutBlock } from './callout'

afterEach(() => {
  cleanup()
})

describe('CalloutBlock', () => {
  it('应默认采用 info 等级并渲染内容', () => {
    render(
      <CalloutBlock
        blockId="co-1"
        props={{ content: '普通提示文本' }}
        depth={1}
      />,
    )

    const el = screen.getByTestId('callout-block')
    expect(el.className).toContain('border-lagoon')
    expect(el.className).toContain('bg-lagoon/5')
    expect(screen.getByText('普通提示文本')).toBeTruthy()
  })

  it('四种等级的语义样式与标题渲染', () => {
    const levels = ['info', 'success', 'warning', 'error'] as const
    const expectedBorders = {
      info: 'border-lagoon',
      success: 'border-kicker',
      warning: 'border-palm',
      error: 'border-destructive',
    }

    for (const lvl of levels) {
      cleanup()
      render(
        <CalloutBlock
          blockId={`co-${lvl}`}
          props={{
            level: lvl,
            title: `${lvl} 标题`,
            content: `${lvl} 内容`,
          }}
          depth={1}
        />,
      )

      const el = screen.getByTestId('callout-block')
      expect(el.className).toContain(expectedBorders[lvl])
      expect(screen.getByText(`${lvl} 标题`)).toBeTruthy()
      expect(screen.getByText(`${lvl} 内容`)).toBeTruthy()
    }
  })
})
