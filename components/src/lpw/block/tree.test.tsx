/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TreeBlock } from './tree'

afterEach(() => {
  cleanup()
})

describe('TreeBlock', () => {
  it('应正确递归渲染树节点及其 note 说明与 title', () => {
    render(
      <TreeBlock
        blockId="tr-1"
        props={{
          title: '项目架构目录',
          nodes: [
            {
              label: 'internal/',
              note: '后端核心',
              children: [
                {
                  label: 'logic/',
                  note: '业务编排',
                },
                {
                  label: 'handler/',
                },
              ],
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('项目架构目录')).toBeTruthy()
    expect(screen.getByText('internal/')).toBeTruthy()
    expect(screen.getByText('(后端核心)')).toBeTruthy()
    expect(screen.getByText('logic/')).toBeTruthy()
    expect(screen.getByText('(业务编排)')).toBeTruthy()
    expect(screen.getByText('handler/')).toBeTruthy()
  })

  it('移动端横向滚动与深层递归截断支持', () => {
    const deepNode = {
      label: 'level-0',
      children: [
        {
          label: 'level-1',
          children: [
            {
              label: 'level-2',
              children: [
                {
                  label: 'level-3',
                  children: [
                    {
                      label: 'level-4',
                      children: [
                        {
                          label: 'level-5',
                          children: [
                            {
                              label: 'level-6',
                              children: [
                                {
                                  label: 'level-7',
                                },
                              ],
                            },
                          ],
                        },
                      ],
                    },
                  ],
                },
              ],
            },
          ],
        },
      ],
    }

    const { container } = render(
      <TreeBlock
        blockId="tr-2"
        props={{
          title: '深度树',
          nodes: [deepNode],
        }}
        depth={1}
      />,
    )

    const block = container.querySelector('[data-testid="tree-block"]')
    expect(block?.className).toContain('overflow-x-auto')
    expect(block?.className).toContain('min-w-0')
    expect(block?.className).toContain('p-3')
    expect(block?.className).toContain('sm:p-5')

    expect(screen.getByText('level-7')).toBeTruthy()
    expect(screen.getByText('level-7').className).toContain('break-all')
  })
})
