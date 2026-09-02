/** @vitest-environment jsdom */
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { groupHistoryDesc } from './group-history'
import { HistoryCard } from './history-card'
import type { Question } from './types'

function q(
  partial: Partial<Question> & Pick<Question, 'id' | 'content'>,
): Question {
  return {
    sessionId: 's1',
    type: 'text',
    groupLabel: '',
    status: 'answered',
    answered: true,
    createdAt: '2026-09-01T00:00:00.000Z',
    ...partial,
  }
}

describe('HistoryCard', () => {
  it('renders groups and answers newest-first', () => {
    const groupedHistory = groupHistoryDesc([
      q({
        id: 'old-a',
        content: 'oldest A question',
        answer: 'old A answer',
        groupLabel: '组 A',
        createdAt: '2026-09-01T01:00:00.000Z',
        answeredAt: '2026-09-01T01:05:00.000Z',
      }),
      q({
        id: 'new-b',
        content: 'newest B question',
        answer: 'new B answer',
        groupLabel: '组 B',
        createdAt: '2026-09-01T03:00:00.000Z',
        answeredAt: '2026-09-01T03:05:00.000Z',
      }),
    ])

    render(<HistoryCard groupedHistory={groupedHistory} />)

    const texts = screen
      .getAllByText(/question|answer|组/)
      .map((el) => el.textContent)
    const joined = texts.join('\n')
    expect(joined.indexOf('组 B')).toBeGreaterThan(-1)
    expect(joined.indexOf('组 B')).toBeLessThan(joined.indexOf('组 A'))
    expect(joined.indexOf('newest B question')).toBeLessThan(
      joined.indexOf('oldest A question'),
    )
  })

  it('shows empty state when there is no history', () => {
    render(<HistoryCard groupedHistory={{}} />)
    expect(screen.getByText('暂无历史记录')).toBeTruthy()
  })
})
