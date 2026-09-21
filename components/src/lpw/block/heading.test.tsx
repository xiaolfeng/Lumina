/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { HeadingBlock } from './heading'

afterEach(() => {
  cleanup()
})

describe('HeadingBlock', () => {
  it('renders heading-1 with 4px lagoon bar and bookmark icon', () => {
    render(
      <HeadingBlock
        blockId="h-1"
        props={{ level: 1, content: '一级标题' }}
        depth={1}
      />,
    )
    const el = screen.getByRole('heading', { level: 1 })
    expect(el.classList.contains('border-l-4')).toBe(true)
    expect(el.classList.contains('border-lagoon')).toBe(true)
    expect(el.querySelector('svg')).toBeTruthy()
  })

  it('level 2 uses thinner bar and smaller icon', () => {
    render(
      <HeadingBlock
        blockId="h-2"
        props={{ level: 2, content: '二级标题' }}
        depth={1}
      />,
    )
    const el = screen.getByRole('heading', { level: 2 })
    expect(el.className).toContain('border-l-[3px]')
  })

  it('level 3 uses soft color', () => {
    render(
      <HeadingBlock
        blockId="h-3"
        props={{ level: 3, content: '三级标题' }}
        depth={1}
      />,
    )
    const el = screen.getByRole('heading', { level: 3 })
    expect(el.classList.contains('text-sea-ink-soft')).toBe(true)
  })

  it('invalid icon falls back to level default', () => {
    render(
      <HeadingBlock
        blockId="h-invalid"
        props={{ level: 1, content: '非法图标标题', icon: 'non-existent' as never }}
        depth={1}
      />,
    )
    const el = screen.getByRole('heading', { level: 1 })
    expect(el.querySelector('svg')).toBeTruthy()
  })

  it('has no bottom border', () => {
    render(
      <HeadingBlock
        blockId="h-noborder"
        props={{ level: 2, content: '无底边标题' }}
        depth={1}
      />,
    )
    const el = screen.getByRole('heading', { level: 2 })
    expect(el.className).not.toContain('border-b')
  })
})
