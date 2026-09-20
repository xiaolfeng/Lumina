/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { CodeBlock } from './code-block'

afterEach(() => {
  cleanup()
})

describe('CodeBlock', () => {
  it('应正确渲染 filename 与 language 顶栏', () => {
    render(
      <CodeBlock
        blockId="cd-1"
        props={{
          filename: 'main.go',
          language: 'go',
          content: 'package main\n\nfunc main() {}',
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('main.go')).toBeTruthy()
    expect(screen.getByText('go')).toBeTruthy()
    expect(screen.getByText('package main')).toBeTruthy()
  })

  it('应在开启 showLineNumbers 时渲染行号并在 highlightLines 命中行加背景', () => {
    const { container } = render(
      <CodeBlock
        blockId="cd-2"
        props={{
          content: 'line 1\nline 2\nline 3',
          showLineNumbers: true,
          highlightLines: [2, 999], // 999 越界
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('1')).toBeTruthy()
    expect(screen.getByText('2')).toBeTruthy()
    expect(screen.getByText('3')).toBeTruthy()

    const line2 = container.querySelector('[data-line-number="2"]')
    expect(line2?.className).toContain('bg-sand')

    const line1 = container.querySelector('[data-line-number="1"]')
    expect(line1?.className).not.toContain('bg-sand')
  })
})
