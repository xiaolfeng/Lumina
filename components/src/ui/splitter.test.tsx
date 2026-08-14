import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, fireEvent } from '@testing-library/react'
import { Splitter, SplitterPanel, SplitterHandle } from './splitter'

/* ── jsdom 未实现 PointerEvent，补齐 polyfill（含坐标）── */
if (typeof PointerEvent === 'undefined') {
  class PointerEventPolyfill extends MouseEvent {
    pointerId: number
    pointerType: string
    isPrimary: boolean
    constructor(type: string, init: PointerEventInit = {}) {
      super(type, init)
      this.pointerId = init.pointerId ?? 0
      this.pointerType = init.pointerType ?? 'mouse'
      this.isPrimary = init.isPrimary ?? true
    }
  }
  ;(globalThis as Record<string, unknown>).PointerEvent = PointerEventPolyfill
}

/* jsdom 未实现 Pointer Capture，补齐 stub */
beforeEach(() => {
  Object.defineProperty(Element.prototype, 'setPointerCapture', {
    value: vi.fn(),
    writable: true,
  })
  Object.defineProperty(Element.prototype, 'releasePointerCapture', {
    value: vi.fn(),
    writable: true,
  })
})

/* 工具：mock 容器矩形（jsdom 默认返回全 0） */
function mockContainerRect(el: HTMLElement, width = 1000, height = 500) {
  vi.spyOn(el, 'getBoundingClientRect').mockReturnValue({
    x: 0,
    y: 0,
    top: 0,
    left: 0,
    right: width,
    bottom: height,
    width,
    height,
    toJSON: () => ({}),
  })
}

function renderSplitter(props?: { direction?: 'horizontal' | 'vertical' }) {
  const { container } = render(
    <Splitter direction={props?.direction} data-testid="splitter">
      <SplitterPanel data-testid="panel-a">左</SplitterPanel>
      <SplitterHandle data-testid="handle-0" />
      <SplitterPanel data-testid="panel-b">右</SplitterPanel>
    </Splitter>
  )
  const splitter = container.querySelector<HTMLElement>('[data-testid="splitter"]')!
  mockContainerRect(splitter)
  const handle = container.querySelector<HTMLElement>('[data-testid="handle-0"]')!
  const panelA = container.querySelector<HTMLElement>('[data-testid="panel-a"]')!
  const panelB = container.querySelector<HTMLElement>('[data-testid="panel-b"]')!
  return { container, splitter, handle, panelA, panelB }
}

function getTrack(splitter: HTMLElement) {
  return splitter.style.getPropertyValue('--splitter-track')
}

describe('Splitter', () => {
  it('renders panels and handle with separator semantics', () => {
    const { container, handle } = renderSplitter()

    expect(container.querySelectorAll('[data-slot="splitter-panel"]')).toHaveLength(2)
    expect(container.querySelectorAll('[data-slot="splitter-handle"]')).toHaveLength(1)
    expect(handle.getAttribute('role')).toBe('separator')
    expect(handle.getAttribute('aria-orientation')).toBe('vertical')
    expect(handle.getAttribute('tabindex')).toBe('0')
  })

  it('initializes default track with last panel absorbing remainder', () => {
    const { splitter } = renderSplitter()

    // 两个面板：第一个 50%，最后一个 1fr
    expect(getTrack(splitter)).toBe('50% var(--splitter-hit-size, 14px) 1fr')
  })

  it('updates ratio on drag via pointer events', () => {
    const { splitter, handle } = renderSplitter()

    fireEvent.pointerDown(handle, { clientX: 500, clientY: 250 })
    fireEvent.pointerMove(handle, { clientX: 600, clientY: 250 })
    fireEvent.pointerUp(handle)

    expect(getTrack(splitter)).toBe('60% var(--splitter-hit-size, 14px) 1fr')
  })

  it('clamps to minSize when dragged to the edge', () => {
    const { splitter, handle } = renderSplitter()

    fireEvent.pointerDown(handle, { clientX: 500, clientY: 250 })
    fireEvent.pointerMove(handle, { clientX: 1, clientY: 250 })
    fireEvent.pointerUp(handle)

    expect(getTrack(splitter)).toBe('12% var(--splitter-hit-size, 14px) 1fr')
  })

  it('supports keyboard fine-tuning with arrow keys', () => {
    const { splitter, handle } = renderSplitter()

    handle.focus()
    fireEvent.keyDown(handle, { key: 'ArrowRight' })
    fireEvent.keyDown(handle, { key: 'ArrowRight' })

    expect(getTrack(splitter)).toBe('52% var(--splitter-hit-size, 14px) 1fr')
  })

  it('supports shift+arrow for larger steps', () => {
    const { splitter, handle } = renderSplitter()

    handle.focus()
    fireEvent.keyDown(handle, { key: 'ArrowLeft', shiftKey: true })

    expect(getTrack(splitter)).toBe('45% var(--splitter-hit-size, 14px) 1fr')
  })

  it('renders vertical direction with horizontal separator orientation', () => {
    const { container, handle, splitter } = renderSplitter({ direction: 'vertical' })

    expect(splitter.getAttribute('data-direction')).toBe('vertical')
    expect(handle.getAttribute('aria-orientation')).toBe('horizontal')
    expect(
      container.querySelector('[data-slot="splitter"]')?.classList.contains(
        'grid-rows-[var(--splitter-track,1fr)]'
      )
    ).toBe(true)
  })

  it('adjusts vertical ratio on vertical drag', () => {
    const { splitter, handle } = renderSplitter({ direction: 'vertical' })

    fireEvent.pointerDown(handle, { clientX: 250, clientY: 250 })
    fireEvent.pointerMove(handle, { clientX: 250, clientY: 300 })
    fireEvent.pointerUp(handle)

    expect(getTrack(splitter)).toBe('60% var(--splitter-hit-size, 14px) 1fr')
  })

  it('restores body cursor after drag ends', () => {
    const { handle } = renderSplitter()

    fireEvent.pointerDown(handle, { clientX: 500, clientY: 250 })
    expect(document.body.style.cursor).toBe('col-resize')

    fireEvent.pointerUp(handle)
    expect(document.body.style.cursor).toBe('')
  })

  it('honors explicit defaultSize and minSize on panels', () => {
    const { container } = render(
      <Splitter data-testid="splitter">
        <SplitterPanel defaultSize={70} minSize={30}>
          左
        </SplitterPanel>
        <SplitterHandle data-testid="handle" />
        <SplitterPanel minSize={20}>右</SplitterPanel>
      </Splitter>
    )
    const splitter = container.querySelector<HTMLElement>('[data-testid="splitter"]')!
    mockContainerRect(splitter)
    const handle = container.querySelector<HTMLElement>('[data-testid="handle"]')!

    expect(getTrack(splitter)).toBe('70% var(--splitter-hit-size, 14px) 1fr')

    // 拖向边缘：左面板被钳制到 minSize=30
    fireEvent.pointerDown(handle, { clientX: 500, clientY: 250 })
    fireEvent.pointerMove(handle, { clientX: 1, clientY: 250 })
    fireEvent.pointerUp(handle)

    expect(getTrack(splitter)).toBe('30% var(--splitter-hit-size, 14px) 1fr')
  })
})
