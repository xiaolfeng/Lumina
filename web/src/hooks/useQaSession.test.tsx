/** @vitest-environment jsdom */
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useQaSession } from './useQaSession'

let capturedOptions: any = null
const mockSendMessage = vi.fn()

vi.mock('./useQaWebSocket', () => ({
  useQaWebSocket: (_hash: string | null, options: any) => {
    capturedOptions = options
    return {
      status: 'connected',
      sendMessage: mockSendMessage,
      connect: vi.fn(),
      disconnect: vi.fn(),
    }
  },
}))

describe('useQaSession', () => {
  it('automatically activates next question supplement when answering previous question', () => {
    const { result } = renderHook(() =>
      useQaSession({ sessionHash: 'test-session' }),
    )

    // 推送两个问题 Q1, Q2
    act(() => {
      capturedOptions.onQuestionPush({
        id: 'q1',
        title: 'Question 1',
        type: 'text',
        supplement: true,
      })
      capturedOptions.onQuestionPush({
        id: 'q2',
        title: 'Question 2',
        type: 'text',
        supplement: true,
      })
    })

    // 分别推送 Q1 和 Q2 的 supplement
    act(() => {
      capturedOptions.onSupplementPush({
        id: 's1',
        target_type: 'question',
        target_id: 'q1',
        content_type: 'preview',
        content: '{"session_id":"s","file_id":"f1"}',
      })
      capturedOptions.onSupplementPush({
        id: 's2',
        target_type: 'question',
        target_id: 'q2',
        content_type: 'preview',
        content: '{"session_id":"s","file_id":"f2"}',
      })
    })

    // 当前在回答 Q1，活跃 supplement 应当是 Q1 的 s1
    expect(result.current.activeSupplement?.id).toBe('s1')

    // 用户作答 Q1
    act(() => {
      result.current.submitAnswer('q1', 'my answer')
    })

    // 作答 Q1 后，Q2 成为首个 pending 问题，Q2 的 s2 必须被自动激活，不能为 null
    expect(result.current.activeSupplement?.id).toBe('s2')

    // 作答 Q2 后，已无 pending 问题，activeSupplement 应为 null
    act(() => {
      result.current.submitAnswer('q2', 'answer 2')
    })
    expect(result.current.activeSupplement).toBeNull()
  })

  it('automatically activates next supplement when skipping a question', () => {
    const { result } = renderHook(() =>
      useQaSession({ sessionHash: 'test-session-skip' }),
    )

    act(() => {
      capturedOptions.onQuestionPush({
        id: 'q1',
        title: 'Q1',
        type: 'text',
      })
      capturedOptions.onQuestionPush({
        id: 'q2',
        title: 'Q2',
        type: 'text',
      })
      capturedOptions.onSupplementPush({
        id: 's2',
        target_type: 'question',
        target_id: 'q2',
        content_type: 'preview',
        content: 'preview 2',
      })
    })

    // Q1 无 supplement
    expect(result.current.activeSupplement).toBeNull()

    // 跳过 Q1
    act(() => {
      result.current.skipQuestion('q1')
    })

    // 自动激活 Q2 的 s2
    expect(result.current.activeSupplement?.id).toBe('s2')
  })

  it('preserves current supplement when a subsequent question is cancelled, and switches when current is cancelled', () => {
    const { result } = renderHook(() =>
      useQaSession({ sessionHash: 'test-session-cancel' }),
    )

    act(() => {
      capturedOptions.onQuestionPush({
        id: 'q1',
        title: 'Q1',
        type: 'text',
      })
      capturedOptions.onQuestionPush({
        id: 'q2',
        title: 'Q2',
        type: 'text',
      })
      capturedOptions.onSupplementPush({
        id: 's1',
        target_type: 'question',
        target_id: 'q1',
        content_type: 'preview',
        content: 'preview 1',
      })
      capturedOptions.onSupplementPush({
        id: 's2',
        target_type: 'question',
        target_id: 'q2',
        content_type: 'preview',
        content: 'preview 2',
      })
    })

    // 当前在 Q1，展示 s1
    expect(result.current.activeSupplement?.id).toBe('s1')

    // 取消后续题目 Q2
    act(() => {
      capturedOptions.onQuestionCancel({
        question_id: 'q2',
        cancel_all: false,
      })
    })

    // 当前 Q1 的 s1 应当保留，不应被粗暴清空
    expect(result.current.activeSupplement?.id).toBe('s1')

    // 取消当前题目 Q1
    act(() => {
      capturedOptions.onQuestionCancel({
        question_id: 'q1',
        cancel_all: false,
      })
    })

    // Q1 已取消，Q2 此前也已取消，此时无待答题目，activeSupplement 应自动为 null
    expect(result.current.activeSupplement).toBeNull()
  })
})
