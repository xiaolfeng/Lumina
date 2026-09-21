/** @vitest-environment jsdom */
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { SessionProgressBar } from './session-progress-bar'
import { computeSessionProgress } from './session-progress'
import { proseQuestion, Markdown } from './primitives'
import {
  Splitter,
  SplitterPanel,
  SplitterHandle,
} from '@lumina/components/ui/splitter'
import type { Question } from './types'

describe('Interact flow integration & regression verification', () => {
  it('Q-01: verifies cancelled items are counted and rendered with dedicated segmented color in progress bar', () => {
    const questions: Question[] = [
      {
        id: 'q1',
        sessionId: 's',
        content: 'Question 1',
        type: 'text',
        groupLabel: '',
        status: 'answered',
        answered: true,
        createdAt: '2026-09-01T00:00:00.000Z',
      },
      {
        id: 'q2',
        sessionId: 's',
        content: 'Question 2',
        type: 'text',
        groupLabel: '',
        status: 'cancelled',
        answered: false,
        createdAt: '2026-09-01T00:00:00.000Z',
      },
      {
        id: 'q3',
        sessionId: 's',
        content: 'Question 3',
        type: 'text',
        groupLabel: '',
        status: 'pending',
        answered: false,
        createdAt: '2026-09-01T00:00:00.000Z',
      },
    ]

    const progress = computeSessionProgress(questions)
    expect(progress).toEqual({
      total: 3,
      answered: 1,
      cancelled: 1,
      remaining: 1,
    })

    const { container } = render(<SessionProgressBar progress={progress} />)

    // 验证文字区包含已答、取消、未答
    expect(screen.getByText('已答')).toBeTruthy()
    expect(screen.getByText('取消')).toBeTruthy()
    expect(screen.getByText('未答')).toBeTruthy()
    expect(screen.getAllByText('1').length).toBeGreaterThanOrEqual(3)

    // 验证进度条分段颜色：已答含 bg-lagoon，取消含 bg-rose-400
    const lagoonSegment = container.querySelector('.bg-lagoon')
    const roseSegment = container.querySelector('.bg-rose-400')
    expect(lagoonSegment).not.toBeNull()
    expect(roseSegment).not.toBeNull()
  })

  it('Q-03: verifies Splitter enforces maxSize=50% and clamps drag at center', () => {
    const { container } = render(
      <Splitter data-testid="splitter">
        <SplitterPanel
          defaultSize={45}
          minSize={30}
          maxSize={50}
          data-testid="panel-left"
        >
          Left Panel
        </SplitterPanel>
        <SplitterHandle data-testid="splitter-handle" />
        <SplitterPanel minSize={30} data-testid="panel-right">
          Right Panel
        </SplitterPanel>
      </Splitter>,
    )

    const splitter = container.querySelector<HTMLElement>(
      '[data-testid="splitter"]',
    )!
    vi.spyOn(splitter, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      top: 0,
      right: 1000,
      bottom: 600,
      width: 1000,
      height: 600,
      x: 0,
      y: 0,
      toJSON: () => {},
    })

    const handle = container.querySelector<HTMLElement>(
      '[data-testid="splitter-handle"]',
    )!

    // 尝试向右拖拽到 80% (800px)
    fireEvent.pointerDown(handle, { clientX: 450, clientY: 300 })
    fireEvent.pointerMove(handle, { clientX: 800, clientY: 300 })
    fireEvent.pointerUp(handle)

    // 验证被 clamp 在 maxSize = 50%
    const track = splitter.style.getPropertyValue('--splitter-track')
    expect(track).toBe('50% var(--splitter-hit-size, 14px) 1fr')
  })

  it('Q-04: verifies Markdown and prose styles enforce word breaking to prevent layout overflow', () => {
    expect(proseQuestion).toContain('break-words')
    expect(proseQuestion).toContain('[overflow-wrap:anywhere]')
    expect(proseQuestion).toContain('[&_code]:break-all')
    expect(proseQuestion).toContain('[&_pre]:overflow-x-auto')

    const longWord =
      'VERY_LONG_UNBROKEN_STRING_THAT_SHOULD_WRAP_AND_NOT_OVERFLOW_CONTAINER'
    const { container } = render(
      <div className={proseQuestion}>
        <Markdown>{`Here is a long word: ${longWord}`}</Markdown>
      </div>,
    )

    const markdownDiv = container.querySelector('[data-slot="markdown"]')
    expect(markdownDiv).not.toBeNull()
    expect(markdownDiv?.classList.contains('min-w-0')).toBe(true)
    expect(markdownDiv?.classList.contains('max-w-full')).toBe(true)
    expect(markdownDiv?.classList.contains('break-words')).toBe(true)
  })
})
