/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { QuestionReview } from './question-review'
import type { Question } from './types'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

function mockQuestion(partial: Partial<Question>): Question {
  return {
    id: 'q-review-1',
    sessionId: 'sess-1',
    type: 'review',
    content: '## 审阅标题\n\n主文案说明内容',
    groupLabel: '热点 2',
    status: 'pending',
    answered: false,
    createdAt: '2026-09-21T00:00:00.000Z',
    ...partial,
  }
}

describe('QuestionReview', () => {
  const sampleSections = [
    {
      id: 'target',
      title: '1. 目标目录与 package',
      content: '在现有根目录 `profiles/` 直接新增 Go 文件',
    },
    {
      id: 'move',
      title: '2. 从 adapter 精确迁出的内容',
      content: '迁出声明类型与 loader',
    },
  ]

  it('正确渲染 sections 列表与段落计数，且主文案不重复', () => {
    const question = mockQuestion({
      config: {
        sections: sampleSections,
      },
    })

    render(
      <QuestionReview
        question={question}
        onSubmit={vi.fn()}
        onSkip={vi.fn()}
        onRequestSupplement={vi.fn()}
      />,
    )

    // 检查计数文案
    expect(screen.getByText('共 2 个审阅段落')).toBeTruthy()

    // 检查各 section 标题已渲染
    expect(screen.getByText('1. 目标目录与 package')).toBeTruthy()
    expect(screen.getByText('2. 从 adapter 精确迁出的内容')).toBeTruthy()

    // 默认全部展开，检查内容
    expect(screen.getByText(/在现有根目录/)).toBeTruthy()
    expect(screen.getByText(/迁出声明类型与 loader/)).toBeTruthy()

    // 主文案应只出现一次（由 QuestionShell 渲染）
    const mainTitles = screen.getAllByText('审阅标题')
    expect(mainTitles).toHaveLength(1)
  })

  it('支持从 JSON 字符串形式的 config 中安全解析 sections', () => {
    const question = mockQuestion({
      config: JSON.stringify({
        sections: sampleSections,
      }) as any,
    })

    render(
      <QuestionReview
        question={question}
        onSubmit={vi.fn()}
        onSkip={vi.fn()}
        onRequestSupplement={vi.fn()}
      />,
    )

    expect(screen.getByText('共 2 个审阅段落')).toBeTruthy()
    expect(screen.getByText('1. 目标目录与 package')).toBeTruthy()
  })

  it('选择 approve 时可直接提交', () => {
    const onSubmit = vi.fn()
    const question = mockQuestion({
      config: {
        sections: sampleSections,
      },
    })

    render(
      <QuestionReview
        question={question}
        onSubmit={onSubmit}
        onSkip={vi.fn()}
        onRequestSupplement={vi.fn()}
      />,
    )

    // 点击批准按钮
    const approveBtn = screen.getByRole('button', { name: /通过|批准/i })
    fireEvent.click(approveBtn)

    // 点击提交按钮
    const submitBtn = screen.getByRole('button', { name: /^提交$/i })
    expect(submitBtn.hasAttribute('disabled')).toBe(false)
    fireEvent.click(submitBtn)

    expect(onSubmit).toHaveBeenCalledWith({
      decision: 'approve',
    })
  })

  it('选择 revise 时支持逐段批注与全局反馈并组装提交', () => {
    const onSubmit = vi.fn()
    const question = mockQuestion({
      config: {
        sections: sampleSections,
      },
    })

    render(
      <QuestionReview
        question={question}
        onSubmit={onSubmit}
        onSkip={vi.fn()}
        onRequestSupplement={vi.fn()}
      />,
    )

    // 点击修改/修订按钮
    const reviseBtn = screen.getByRole('button', { name: /修改|修订/i })
    fireEvent.click(reviseBtn)

    // 未填写任何反馈时提交按钮禁用
    const submitBtn = screen.getByRole('button', { name: /^提交$/i })
    expect(submitBtn.hasAttribute('disabled')).toBe(true)

    // 填写某个 section 的批注
    const sectionInput = screen.getByPlaceholderText(
      /对「1. 目标目录与 package」的修改批注/,
    )
    fireEvent.change(sectionInput, {
      target: { value: '建议加上 Makefile 目标' },
    })

    // 此时已满足至少有一项反馈，按钮可用
    expect(submitBtn.hasAttribute('disabled')).toBe(false)

    // 再填写总体意见
    const globalInput = screen.getByPlaceholderText(/总体修改意见/)
    fireEvent.change(globalInput, { target: { value: '整体满意，微调一下' } })

    fireEvent.click(submitBtn)

    expect(onSubmit).toHaveBeenCalledWith({
      decision: 'revise',
      feedback: '整体满意，微调一下',
      annotations: [
        {
          sectionId: 'target',
          content: '建议加上 Makefile 目标',
        },
      ],
    })
  })

  it('支持全部折叠与全部展开切换', () => {
    const question = mockQuestion({
      config: {
        sections: sampleSections,
      },
    })

    render(
      <QuestionReview
        question={question}
        onSubmit={vi.fn()}
        onSkip={vi.fn()}
        onRequestSupplement={vi.fn()}
      />,
    )

    const toggleAllBtn = screen.getByRole('button', { name: /全部收起/i })
    fireEvent.click(toggleAllBtn)

    // 收起后内容不可见
    expect(screen.queryByText(/在现有根目录/)).toBeNull()
    expect(screen.getByRole('button', { name: /全部展开/i })).toBeTruthy()

    // 再次点击展开
    fireEvent.click(screen.getByRole('button', { name: /全部展开/i }))
    expect(screen.getByText(/在现有根目录/)).toBeTruthy()
  })
})
