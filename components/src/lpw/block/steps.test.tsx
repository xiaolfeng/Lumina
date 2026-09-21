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

    const item1 = container.querySelector('[data-testid="step-item-1"]')
    expect(item1?.innerHTML).toContain('border-lagoon')
    expect(item1?.innerHTML).toContain('ring-2') // current 高亮

    const item2 = container.querySelector('[data-testid="step-item-2"]')
    expect(item2?.innerHTML).toContain('border-line')

    const item3 = container.querySelector('[data-testid="step-item-3"]')
    expect(item3?.innerHTML).toContain('border-destructive')
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
})
