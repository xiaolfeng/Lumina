/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ProgressBlock } from './progress'

afterEach(() => {
  cleanup()
})

describe('ProgressBlock', () => {
  it('应正确展示三段推断色与条宽 style', () => {
    render(
      <ProgressBlock
        blockId="pg-1"
        props={{
          title: '交付进度',
          items: [
            { label: '第一批', value: 100 },
            { label: '第二批', value: 50 },
            { label: '第三批', value: 0 },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('交付进度')).toBeTruthy()
    expect(screen.getByText('100%')).toBeTruthy()
    expect(screen.getByText('50%')).toBeTruthy()
    expect(screen.getByText('0%')).toBeTruthy()

    const bar0 = screen.getByTestId('progress-bar-0')
    expect(bar0.className).toContain('bg-kicker') // 100 自动推断为 finish
    expect(bar0.style.width).toBe('100%')

    const bar1 = screen.getByTestId('progress-bar-1')
    expect(bar1.className).toContain('bg-lagoon') // 50 自动推断为 process
    expect(bar1.style.width).toBe('50%')

    const bar2 = screen.getByTestId('progress-bar-2')
    expect(bar2.className).toContain('bg-line') // 0 自动推断为 wait
    expect(bar2.style.width).toBe('0%')
  })

  it('显式设置 status 时应优先覆盖推断色', () => {
    render(
      <ProgressBlock
        blockId="pg-2"
        props={{
          items: [{ label: '失败任务', value: 60, status: 'error' }],
        }}
        depth={1}
      />,
    )

    const bar = screen.getByTestId('progress-bar-0')
    expect(bar.className).toContain('bg-destructive')
  })
})
