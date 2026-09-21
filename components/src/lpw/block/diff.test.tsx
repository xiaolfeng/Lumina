/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { DiffBlock } from './diff'

afterEach(() => {
  cleanup()
})

describe('DiffBlock', () => {
  it('应正确渲染顶栏与并排/统一视图切换按钮，默认 splitView 为 true', () => {
    render(
      <DiffBlock
        blockId="df-1"
        props={{
          filename: 'router.go',
          language: 'go',
          oldCode: 'const x = 1',
          newCode: 'const x = 2',
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('router.go')).toBeTruthy()
    expect(screen.getByText('go')).toBeTruthy()

    const toggleBtn = screen.getByRole('button', { name: '切换为统一视图' })
    expect(toggleBtn).toBeTruthy()

    fireEvent.click(toggleBtn)
    expect(screen.getByRole('button', { name: '切换为并排视图' })).toBeTruthy()
  })

  it('代码完全相同时渲染无差异占位提示', () => {
    render(
      <DiffBlock
        blockId="df-same"
        props={{
          oldCode: 'hello world',
          newCode: 'hello world',
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('文件内容一致，无差异')).toBeTruthy()
  })

  it('Q-09: 移动端视口宽度小于 640 时默认 splitView 为 false (统一视图)', () => {
    const originalInnerWidth = window.innerWidth
    try {
      window.innerWidth = 500
      render(
        <DiffBlock
          blockId="df-mobile"
          props={{
            filename: 'mobile.go',
            oldCode: 'const x = 1',
            newCode: 'const x = 2',
          }}
          depth={1}
        />,
      )

      // 移动端默认是统一视图，按钮文字应当是“切换为并排视图”
      expect(screen.getByRole('button', { name: '切换为并排视图' })).toBeTruthy()
    } finally {
      window.innerWidth = originalInnerWidth
    }
  })

  it('Q-17 & Q-25: 顶栏长文件名截断、按钮 shrink-0，以及视图切换图标 (Columns / Rows)', () => {
    render(
      <DiffBlock
        blockId="df-long"
        props={{
          filename: 'very-long-directory-path-to-source-code-file.controller.ts',
          language: 'typescript',
          oldCode: 'const a = 1',
          newCode: 'const a = 2',
        }}
        depth={1}
      />,
    )

    const filenameEl = screen.getByText('very-long-directory-path-to-source-code-file.controller.ts')
    expect(filenameEl.className).toContain('truncate')
    expect(filenameEl.className).toContain('max-w-[160px]')
    expect(filenameEl.className).toContain('sm:max-w-sm')
    expect(filenameEl.getAttribute('title')).toBe('very-long-directory-path-to-source-code-file.controller.ts')

    const button = screen.getByRole('button', { name: '切换为统一视图' })
    expect(button.className).toContain('shrink-0')

    // 默认 splitView=true，切换为统一视图按钮前展示 Rows 图标，或者根据当前视图展示 Columns/Rows
    const icon = button.querySelector('svg')
    expect(icon).toBeTruthy()
  })

  it('Q-25: 外部 props.splitView 变化响应同步', () => {
    const { rerender } = render(
      <DiffBlock
        blockId="df-sync"
        props={{
          filename: 'sync.go',
          oldCode: 'const x = 1',
          newCode: 'const x = 2',
          splitView: true,
        }}
        depth={1}
      />,
    )

    expect(screen.getByRole('button', { name: '切换为统一视图' })).toBeTruthy()

    rerender(
      <DiffBlock
        blockId="df-sync"
        props={{
          filename: 'sync.go',
          oldCode: 'const x = 1',
          newCode: 'const x = 2',
          splitView: false,
        }}
        depth={1}
      />,
    )

    expect(screen.getByRole('button', { name: '切换为并排视图' })).toBeTruthy()
  })

  it('Q-05: 移除 || true 恒真项，无 filename 和 language 时不展示顶栏', () => {
    render(
      <DiffBlock
        blockId="df-no-header"
        props={{
          oldCode: 'const x = 1',
          newCode: 'const x = 2',
        }}
        depth={1}
      />,
    )

    // 不应渲染顶栏切换按钮与边框栏
    expect(screen.queryByRole('button', { name: /切换为/ })).toBeNull()
  })
})
