/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ListBlock } from './list'

afterEach(() => {
  cleanup()
})

describe('ListBlock', () => {
  it('应默认渲染无序列表', () => {
    const { container } = render(
      <ListBlock
        blockId="l-1"
        props={{
          items: [{ content: '无序 1' }, { content: '无序 **加粗**' }],
        }}
        depth={1}
      />,
    )

    expect(container.querySelector('ul.list-disc')).toBeTruthy()
    expect(screen.getByText('无序 1')).toBeTruthy()
    expect(screen.getByText('加粗')).toBeTruthy()
  })

  it('应正确渲染有序列表', () => {
    const { container } = render(
      <ListBlock
        blockId="l-2"
        props={{
          style: 'ordered',
          items: [{ content: '有序 1' }],
        }}
        depth={1}
      />,
    )

    expect(container.querySelector('ol.list-decimal')).toBeTruthy()
    expect(screen.getByText('有序 1')).toBeTruthy()
  })

  it('应正确渲染 check 勾选列表与只读状态标记', () => {
    render(
      <ListBlock
        blockId="l-3"
        props={{
          style: 'check',
          items: [
            { content: '已完成项', checked: true },
            { content: '未完成项', checked: false },
          ],
        }}
        depth={1}
      />,
    )

    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes[0].getAttribute('aria-checked')).toBe('true')
    expect(checkboxes[0].textContent).toBe('✓')
    expect(checkboxes[1].getAttribute('aria-checked')).toBe('false')
    expect(checkboxes[1].textContent).toBe('□')
  })
})
