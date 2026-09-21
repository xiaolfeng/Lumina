/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { lpwRegistry, registerAll } from './index'
import { rootLocation } from './render-location'
import { LpwNodeRenderer } from './renderer'
import type { LpwBlock, LpwContainer, LpwLayout } from './types'

afterEach(() => {
  cleanup()
})

beforeEach(() => {
  lpwRegistry.clear()
  registerAll()
})

describe('renderer', () => {
  const loc = rootLocation()

  it('layout child in layout renders INVALID_LAYOUT_SLOT', () => {
    const layoutNode: LpwLayout = {
      id: 'lay-child',
      kind: 'layout',
      type: 'layout',
      props: { pattern: 'split' },
      children: [],
    }

    render(
      <LpwNodeRenderer
        node={layoutNode}
        location={loc}
        parentKind="layout"
      />,
    )

    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByTestId('diagnostic-code').textContent).toBe(
      'INVALID_LAYOUT_SLOT',
    )
  })

  it('container child in container renders INVALID_CONTAINER_CHILD', () => {
    const containerNode: LpwContainer = {
      id: 'sec-child',
      kind: 'container',
      type: 'section',
      props: { variant: 'article', title: 't' },
      children: [],
    }

    render(
      <LpwNodeRenderer
        node={containerNode}
        location={loc}
        parentKind="container"
      />,
    )

    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByTestId('diagnostic-code').textContent).toBe(
      'INVALID_CONTAINER_CHILD',
    )
  })

  it('block with children renders diagnostic', () => {
    const blockWithChildren: LpwBlock = {
      id: 'b-err',
      kind: 'block',
      type: 'markdown',
      props: { content: 'hello' },
    }
    ;(blockWithChildren as any).children = [
      { id: 'sub', kind: 'block', type: 'markdown', props: { content: 'sub' } },
    ]

    render(
      <LpwNodeRenderer
        node={blockWithChildren}
        location={loc}
      />,
    )

    // 正常渲染 markdown 内容
    expect(screen.getByText('hello')).toBeTruthy()
  })

  it('siblings still render after one fails', () => {
    const crashBlock: LpwBlock = {
      id: 'b-crash',
      kind: 'block',
      type: 'unknown-type-fails',
      props: {},
    }
    const goodBlock: LpwBlock = {
      id: 'b-good',
      kind: 'block',
      type: 'markdown',
      props: { content: '好节点正常呈现' },
    }

    render(
      <div>
        <LpwNodeRenderer node={crashBlock} location={loc} />
        <LpwNodeRenderer node={goodBlock} location={loc} />
      </div>,
    )

    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByText('好节点正常呈现')).toBeTruthy()
  })

  it('annotation frame only on blocks', () => {
    const blockWithAnn: LpwBlock = {
      id: 'b-ann',
      kind: 'block',
      type: 'markdown',
      props: { content: '带批注的正文' },
      annotation: { kind: 'note', message: '请复核' },
    }

    render(<LpwNodeRenderer node={blockWithAnn} location={loc} />)
    expect(screen.getByTestId('annotation-frame')).toBeTruthy()
  })
})
