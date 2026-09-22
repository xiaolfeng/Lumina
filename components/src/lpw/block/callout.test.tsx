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

  it('Q-03 未受控 level 安全回退至 info', () => {
    render(
      <CalloutBlock
        blockId="co-invalid"
        props={{
          level: 'unknown-level' as never,
          content: '回退提示',
        }}
        depth={1}
      />,
    )

    const el = screen.getByTestId('callout-block')
    expect(el.getAttribute('data-level')).toBe('info')
    expect(el.className).toContain('border-lagoon')
  })

  it('Q-08 内边距为 p-4 sm:p-5 且文本容器具有长词换行保护', () => {
    const { container } = render(
      <CalloutBlock
        blockId="co-padding"
        props={{ content: '换行与内边距保护' }}
        depth={1}
      />,
    )

    const root = container.firstElementChild as HTMLElement
    expect(root.className).toContain('p-4')
    expect(root.className).toContain('sm:p-5')
    expect(root.className.split(/\s+/)).not.toContain('p-5')

    const contentWrapper = screen.getByText('换行与内边距保护').closest('div')
    expect(contentWrapper?.className).toContain('break-words')
    expect(contentWrapper?.className).toContain('[overflow-wrap:anywhere]')
  })

  it('Q-09 移除 border-double 并保持单侧加粗', () => {
    const { container } = render(
      <CalloutBlock
        blockId="co-border"
        props={{ content: '边框样式检查' }}
        depth={1}
      />,
    )

    const root = container.firstElementChild as HTMLElement
    expect(root.className).not.toContain('border-double')
    expect(root.className).toContain('border-solid')
    expect(root.className).toContain('border-l-4')
  })

  it('Q-12 role 属性在 error 等级为 alert，其余为 status', () => {
    const { rerender } = render(
      <CalloutBlock
        blockId="co-role-warn"
        props={{ level: 'warning', content: '警告' }}
        depth={1}
      />,
    )
    expect(screen.getByTestId('callout-block').getAttribute('role')).toBe('status')

    rerender(
      <CalloutBlock
        blockId="co-role-err"
        props={{ level: 'error', content: '错误' }}
        depth={1}
      />,
    )
    expect(screen.getByTestId('callout-block').getAttribute('role')).toBe('alert')
  })

  it('Q-14 根 DOM 节点显式补齐 id', () => {
    const { container } = render(
      <CalloutBlock
        blockId="co-id-1"
        props={{ content: '测试 id' }}
        depth={1}
      />,
    )
    expect(container.firstElementChild?.getAttribute('id')).toBe('co-id-1')
  })

  it('Q-30 分类标签前渲染 Lucide 图标', () => {
    const { container } = render(
      <CalloutBlock
        blockId="co-icon"
        props={{ level: 'info', content: '带图标提示' }}
        depth={1}
      />,
    )
    const icon = container.querySelector('svg')
    expect(icon).toBeTruthy()
    expect(icon?.getAttribute('aria-hidden')).toBe('true')
  })
})
