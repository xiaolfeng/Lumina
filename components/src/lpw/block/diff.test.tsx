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
})
