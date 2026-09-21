/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { StepsBlock } from './steps'

afterEach(() => {
  cleanup()
})

describe('StepsBlock', () => {
  it('四种状态色类名与 current 高亮正确应用', () => {
    const { container } = render(
      <StepsBlock
        blockId="st-1"
        props={{
          current: 1,
          items: [
            { title: '第一步', status: 'finish' },
            { title: '第二步', status: 'process', desc: '进行中说明' },
            { title: '第三步', status: 'wait' },
            { title: '第四步', status: 'error' },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('第一步')).toBeTruthy()
    expect(screen.getByText('第二步')).toBeTruthy()
    expect(screen.getByText('进行中说明')).toBeTruthy()

    // 状态类名检查
    const item0 = container.querySelector('[data-testid="step-item-0"]')
    expect(item0?.innerHTML).toContain('border-kicker')
    // 矢量图标 Check
    expect(item0?.querySelector('svg')).toBeTruthy()

    const item1 = container.querySelector('[data-testid="step-item-1"]')
    expect(item1?.innerHTML).toContain('border-lagoon')
    expect(item1?.innerHTML).toContain('ring-2') // current 高亮

    const item2 = container.querySelector('[data-testid="step-item-2"]')
    expect(item2?.innerHTML).toContain('border-line')

    const item3 = container.querySelector('[data-testid="step-item-3"]')
    expect(item3?.innerHTML).toContain('border-destructive')
    // 矢量图标 AlertCircle
    expect(item3?.querySelector('svg')).toBeTruthy()
  })

  it('缺省 current 时不高亮任何项', () => {
    const { container } = render(
      <StepsBlock
        blockId="st-2"
        props={{
          items: [{ title: '普通步骤' }],
        }}
        depth={1}
      />,
    )

    const item0 = container.querySelector('[data-testid="step-item-0"]')
    expect(item0?.innerHTML).not.toContain('ring-2')
  })

  it('Q-06: 未显式传 status 时的状态推导（idx < current 为 finish, idx === current 为 process, idx > current 为 wait）', () => {
    const { container } = render(
      <StepsBlock
        blockId="st-infer"
        props={{
          current: 1,
          items: [
            { title: '推导完成步骤' },
            { title: '推导进行中步骤' },
            { title: '推导等待步骤' },
          ],
        }}
        depth={1}
      />,
    )

    const item0 = container.querySelector('[data-testid="step-item-0"]')
    const item1 = container.querySelector('[data-testid="step-item-1"]')
    const item2 = container.querySelector('[data-testid="step-item-2"]')

    // idx=0 < current(1) => finish => kicker 色 + Check 图标
    expect(item0?.innerHTML).toContain('border-kicker')
    expect(item0?.querySelector('svg')).toBeTruthy()

    // idx=1 === current(1) => process => lagoon 色 + ring-2
    expect(item1?.innerHTML).toContain('border-lagoon')
    expect(item1?.innerHTML).toContain('ring-2')

    // idx=2 > current(1) => wait => border-line
    expect(item2?.innerHTML).toContain('border-line')
    expect(item2?.querySelector('svg')).toBeNull()
    expect(item2?.textContent).toContain('3')
  })

  it('Q-06: 显式 status 优先级高于 current 推导，不冲突叠加', () => {
    const { container } = render(
      <StepsBlock
        blockId="st-override"
        props={{
          current: 2,
          items: [
            { title: '强制等待', status: 'wait' },
            { title: '强制报错', status: 'error' },
            { title: '当前步骤' },
          ],
        }}
        depth={1}
      />,
    )

    const item0 = container.querySelector('[data-testid="step-item-0"]')
    const item1 = container.querySelector('[data-testid="step-item-1"]')

    // idx=0 < current(2)，但显式声明 status: 'wait'，不应被强制推导为 finish
    expect(item0?.innerHTML).toContain('border-line')
    expect(item0?.innerHTML).not.toContain('border-kicker')

    // idx=1 < current(2)，但显式声明 status: 'error'
    expect(item1?.innerHTML).toContain('border-destructive')
  })

  it('Q-14: 容器类名响应式布局与移动端连接线视觉元素', () => {
    const { container } = render(
      <StepsBlock
        blockId="st-responsive"
        props={{
          current: 0,
          items: [
            { title: '步骤一' },
            { title: '步骤二' },
          ],
        }}
        depth={1}
      />,
    )

    const ol = container.querySelector('ol')
    expect(ol?.className).toContain('my-6 flex flex-col sm:flex-row sm:flex-wrap items-start gap-4 sm:gap-8 text-xs font-sans')

    // 移动端全宽类名
    const item0 = container.querySelector('[data-testid="step-item-0"]')
    expect(item0?.className).toContain('w-full sm:w-auto')
  })
})
