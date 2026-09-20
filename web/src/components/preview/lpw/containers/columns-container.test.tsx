/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../lpw-registry'
import { ColumnsContainer } from './columns-container'

afterEach(() => {
  cleanup()
})

describe('ColumnsContainer', () => {
  it('四种 ratio 对应的网格类名正确映射，且基础类包含 grid-cols-1', () => {
    lpwRegistry.register('col-item', false, ({ blockId }) => (
      <div>内容 {blockId}</div>
    ))

    const ratios = ['1:1', '1:2', '2:1', '1:1:1'] as const
    const expectedClasses = {
      '1:1': 'md:grid-cols-2',
      '1:2': 'md:grid-cols-[1fr_2fr]',
      '2:1': 'md:grid-cols-[2fr_1fr]',
      '1:1:1': 'md:grid-cols-3',
    }

    for (const r of ratios) {
      cleanup()
      const { container } = render(
        <ColumnsContainer
          blockId="cols-test"
          props={{ ratio: r }}
          depth={1}
          childrenBlocks={[
            { id: 'c1', type: 'col-item', props: {} },
            { id: 'c2', type: 'col-item', props: {} },
          ]}
        />,
      )

      const box = container.querySelector('[data-testid="columns-container"]')
      expect(box?.className).toContain('grid-cols-1')
      expect(box?.className).toContain(expectedClasses[r])
      expect(screen.getByText('内容 c1')).toBeTruthy()
      expect(screen.getByText('内容 c2')).toBeTruthy()
    }
  })
})
