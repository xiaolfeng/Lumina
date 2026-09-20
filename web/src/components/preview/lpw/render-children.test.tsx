/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from './lpw-registry'
import { renderChildren } from './render-children'

afterEach(() => {
  cleanup()
})

describe('renderChildren', () => {
  it('应遍历渲染子块并将 depth 递增 1 传递给下级', () => {
    lpwRegistry.register('depth-tester', false, ({ depth, blockId }) => (
      <div data-testid={`depth-node-${blockId}`} data-depth={depth}>
        depth: {depth}
      </div>
    ))

    const children = [
      { id: 'c1', type: 'depth-tester', props: {} },
      { id: 'c2', type: 'depth-tester', props: {} },
    ]

    render(<div>{renderChildren(children, 1)}</div>)

    const n1 = screen.getByTestId('depth-node-c1')
    expect(n1.getAttribute('data-depth')).toBe('2')

    const n2 = screen.getByTestId('depth-node-c2')
    expect(n2.getAttribute('data-depth')).toBe('2')
  })

  it('传入 undefined 时应返回 undefined', () => {
    expect(renderChildren(undefined, 1)).toBeUndefined()
  })
})
