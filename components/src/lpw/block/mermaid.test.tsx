/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MermaidBlock } from './mermaid'

afterEach(() => {
  cleanup()
})

describe('MermaidBlock', () => {
  it('应包装为 markdown mermaid fence 代码块并渲染，包含 caption', () => {
    render(
      <MermaidBlock
        blockId="mm-1"
        props={{
          content: 'flowchart LR\n  A-->B',
          caption: '图 2. 流程拓扑',
        }}
        depth={1}
      />,
    )

    expect(screen.getByTestId('mermaid-block')).toBeTruthy()
    expect(screen.getByText('图 2. 流程拓扑')).toBeTruthy()
  })

  it('默认页面级宽度应为 80%，嵌套在 layout/container 内时应为 100%，且支持自定义 width', () => {
    const { rerender } = render(
      <MermaidBlock
        blockId="mm-top"
        props={{
          content: 'flowchart LR\n  A-->B',
        }}
        location={{ jsonPath: '/content/0', idPath: [] }}
      />,
    )
    const fig = screen.getByTestId('mermaid-block')
    expect(fig.style.width).toBe('80%')

    rerender(
      <MermaidBlock
        blockId="mm-nested"
        props={{
          content: 'flowchart LR\n  A-->B',
        }}
        location={{ jsonPath: '/content/0/children/0', idPath: ['container-1'] }}
      />,
    )
    expect(fig.style.width).toBe('100%')

    rerender(
      <MermaidBlock
        blockId="mm-custom"
        props={{
          content: 'flowchart LR\n  A-->B',
          width: '60%',
        }}
        location={{ jsonPath: '/content/0', idPath: [] }}
      />,
    )
    expect(fig.style.width).toBe('60%')
  })
})
