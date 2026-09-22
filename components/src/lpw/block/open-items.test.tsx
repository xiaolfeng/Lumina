/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { OpenItemsBlock } from './open-items'

afterEach(() => {
  cleanup()
})

describe('OpenItemsBlock', () => {
  it('应正确渲染克制边框与默认未决事项标题', () => {
    const { container } = render(
      <OpenItemsBlock
        blockId="oi-1"
        props={{
          items: [
            {
              title: '待联调网关',
              detail: '等待平台下发证书',
              owner: '李四',
              due: '下周五前',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('未决事项')).toBeTruthy()
    expect(screen.getByText('待联调网关')).toBeTruthy()
    expect(screen.getByText('等待平台下发证书')).toBeTruthy()
    expect(screen.getByText(/负责人: 李四/)).toBeTruthy()
    expect(screen.getByText(/截止: 下周五前/)).toBeTruthy()

    const box = container.querySelector('[data-testid="open-items-block"]')
    expect(box?.className).toContain('border-line')
  })

  it('缺省 owner 与 due 时省略右侧槽位', () => {
    render(
      <OpenItemsBlock
        blockId="oi-2"
        props={{
          title: '自定义任务',
          items: [{ title: '普通待办' }],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('自定义任务')).toBeTruthy()
    expect(screen.getByText('普通待办')).toBeTruthy()
    expect(screen.queryByText(/负责人/)).toBeNull()
  })

  it('元信息与标题容器换行支持及 Lucide 图标与空态渲染', () => {
    const { container, rerender } = render(
      <OpenItemsBlock
        blockId="oi-3"
        props={{
          items: [
            {
              title: '长标题待办事项',
              owner: '张三',
              due: '2026-12-31',
            },
          ],
        }}
        depth={1}
      />,
    )

    // 检查图标渲染
    expect(container.querySelector('.lucide-circle')).toBeTruthy()
    expect(container.querySelector('.lucide-user')).toBeTruthy()
    expect(container.querySelector('.lucide-calendar')).toBeTruthy()

    // 检查空态
    rerender(
      <OpenItemsBlock
        blockId="oi-empty"
        props={{
          items: [],
        }}
        depth={1}
      />,
    )
    expect(screen.getByText('暂无待决事项')).toBeTruthy()
  })
})
