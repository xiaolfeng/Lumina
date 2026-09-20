/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { HeadingBlock } from './heading-block'

afterEach(() => {
  cleanup()
})

describe('HeadingBlock', () => {
  it('应默认采用 level 2 渲染 h2 且包含 blockId 作为 id', () => {
    render(
      <HeadingBlock
        blockId="h-default"
        props={{ content: '默认标题' }}
        depth={1}
      />,
    )

    const el = screen.getByRole('heading', { level: 2 })
    expect(el).toBeTruthy()
    expect(el.id).toBe('h-default')
    expect(el.textContent).toBe('默认标题')
  })

  it('显式指定 level 1 与 3 时应分别渲染 h1 与 h3', () => {
    render(
      <HeadingBlock
        blockId="h-1"
        props={{ level: 1, content: '一级标题' }}
        depth={1}
      />,
    )
    expect(screen.getByRole('heading', { level: 1 }).id).toBe('h-1')

    render(
      <HeadingBlock
        blockId="h-3"
        props={{ level: 3, content: '三级标题' }}
        depth={1}
      />,
    )
    expect(screen.getByRole('heading', { level: 3 }).id).toBe('h-3')
  })
})
