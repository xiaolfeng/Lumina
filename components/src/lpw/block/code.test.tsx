/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { CodeBlock } from './code'

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

  it('Q-17 & Q-25: 顶栏长文件名防挤爆并包含 FileCode 图标', () => {
    const { container } = render(
      <CodeBlock
        blockId="cd-long-filename"
        props={{
          filename: 'very-long-path-to-deeply-nested-configuration-file.component.tsx',
          language: 'tsx',
          content: 'const a = 1',
        }}
        depth={1}
      />,
    )

    const filenameEl = screen.getByText('very-long-path-to-deeply-nested-configuration-file.component.tsx')
    expect(filenameEl.className).toContain('truncate')
    expect(filenameEl.className).toContain('max-w-[200px]')
    expect(filenameEl.className).toContain('sm:max-w-md')
    expect(filenameEl.getAttribute('title')).toBe('very-long-path-to-deeply-nested-configuration-file.component.tsx')

    // FileCode 图标位于文件名左侧
    const icon = container.querySelector('svg.lucide-file-code, svg')
    expect(icon).toBeTruthy()
  })

  it('Q-03: 行号宽度支持自适应，大行号切换为 w-10', () => {
    const smallContent = 'a\nb'
    const { container: smallContainer } = render(
      <CodeBlock
        blockId="cd-small"
        props={{
          content: smallContent,
          showLineNumbers: true,
        }}
        depth={1}
      />,
    )

    const smallLineNo = smallContainer.querySelector('[data-line-number="1"] > span')
    expect(smallLineNo?.className).toContain('w-8')

    cleanup()

    const largeContent = Array.from({ length: 1005 }, (_, i) => `line ${i + 1}`).join('\n')
    const { container: largeContainer } = render(
      <CodeBlock
        blockId="cd-large"
        props={{
          content: largeContent,
          showLineNumbers: true,
        }}
        depth={1}
      />,
    )

    const largeLineNo = largeContainer.querySelector('[data-line-number="1000"] > span')
    expect(largeLineNo?.className).toContain('w-10')
  })

  it('Q-04: 代码行包含 min-w-full w-fit 与 items-start', () => {
    const { container } = render(
      <CodeBlock
        blockId="cd-line-style"
        props={{
          content: 'console.log("hello world");',
          highlightLines: [1],
        }}
        depth={1}
      />,
    )

    const line = container.querySelector('[data-line-number="1"]')
    expect(line?.className).toContain('min-w-full')
    expect(line?.className).toContain('w-fit')
    expect(line?.className).toContain('items-start')
  })
})
