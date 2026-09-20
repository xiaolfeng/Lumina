/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../lpw-registry'
import { DetailsContainer } from './details-container'

afterEach(() => {
  cleanup()
})

describe('DetailsContainer', () => {
  it('应正确渲染原生 details/summary，defaultOpen 为 true 时初始包含 open 属性', () => {
    lpwRegistry.register('det-child', false, () => <div>折叠内详细代码</div>)

    const { rerender } = render(
      <DetailsContainer
        blockId="det-1"
        props={{ summary: '更多配置', defaultOpen: true }}
        depth={1}
        childrenBlocks={[{ id: 'cd-1', type: 'det-child', props: {} }]}
      />,
    )

    const detailsEl = screen.getByTestId('details-container')
    expect(detailsEl.hasAttribute('open')).toBe(true)
    expect(screen.getByText('更多配置')).toBeTruthy()
    expect(screen.getByText('折叠内详细代码')).toBeTruthy()

    rerender(
      <DetailsContainer
        blockId="det-2"
        props={{ summary: '收起状态', defaultOpen: false }}
        depth={1}
        childrenBlocks={[{ id: 'cd-2', type: 'det-child', props: {} }]}
      />,
    )
    expect(detailsEl.hasAttribute('open')).toBe(false)
  })
})
