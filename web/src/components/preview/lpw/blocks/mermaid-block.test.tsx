/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MermaidBlock } from './mermaid-block'

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
})
