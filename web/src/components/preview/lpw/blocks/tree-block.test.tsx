/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TreeBlock } from './tree-block'

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
})
