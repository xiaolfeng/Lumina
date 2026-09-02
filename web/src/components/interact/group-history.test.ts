import { describe, expect, it } from 'vitest'

import {
  groupHistoryDesc,
  historyEventTime,
  isHistoryQuestion,
} from './group-history'
import type { Question } from './types'

function q(partial: Partial<Question> & Pick<Question, 'id'>): Question {
  return {
    sessionId: 's1',
    content: partial.content ?? partial.id,
    type: 'text',
    groupLabel: '',
    status: 'answered',
    answered: true,
    createdAt: '2026-09-01T00:00:00.000Z',
    ...partial,
  }
}

describe('isHistoryQuestion', () => {
  it('keeps answered and cancelled, drops pending and skipped', () => {
    expect(
      isHistoryQuestion(q({ id: 'a', status: 'answered', answered: true })),
    ).toBe(true)
    expect(
      isHistoryQuestion(q({ id: 'c', status: 'cancelled', answered: false })),
    ).toBe(true)
    expect(
      isHistoryQuestion(q({ id: 'p', status: 'pending', answered: false })),
    ).toBe(false)
    expect(
      isHistoryQuestion(q({ id: 'k', status: 'skipped', answered: false })),
    ).toBe(false)
  })
})

describe('historyEventTime', () => {
  it('prefers answeredAt over createdAt', () => {
    const item = q({
      id: 'a',
      createdAt: '2026-09-01T00:00:00.000Z',
      answeredAt: '2026-09-02T12:00:00.000Z',
    })
    expect(historyEventTime(item)).toBe(Date.parse('2026-09-02T12:00:00.000Z'))
  })
})

describe('groupHistoryDesc', () => {
  it('puts the newest group first even when older questions arrived earlier', () => {
    const grouped = groupHistoryDesc([
      q({
        id: 'old-a',
        content: 'oldest A',
        groupLabel: '组 A',
        createdAt: '2026-09-01T01:00:00.000Z',
        answeredAt: '2026-09-01T01:05:00.000Z',
      }),
      q({
        id: 'mid-a',
        content: 'newer A',
        groupLabel: '组 A',
        createdAt: '2026-09-01T02:00:00.000Z',
        answeredAt: '2026-09-01T02:05:00.000Z',
      }),
      q({
        id: 'new-b',
        content: 'newest B',
        groupLabel: '组 B',
        createdAt: '2026-09-01T03:00:00.000Z',
        answeredAt: '2026-09-01T03:05:00.000Z',
      }),
    ])

    expect(Object.keys(grouped)).toEqual(['组 B', '组 A'])
    expect(grouped['组 B'].map((item) => item.id)).toEqual(['new-b'])
    expect(grouped['组 A'].map((item) => item.id)).toEqual(['mid-a', 'old-a'])
  })

  it('lifts a just-answered question to the top via answeredAt', () => {
    const grouped = groupHistoryDesc([
      q({
        id: 'old',
        content: 'old answered',
        groupLabel: '未分组',
        createdAt: '2026-09-01T08:00:00.000Z',
        answeredAt: '2026-09-01T08:01:00.000Z',
      }),
      q({
        id: 'fresh',
        content: 'just answered',
        groupLabel: '未分组',
        createdAt: '2026-09-01T07:00:00.000Z',
        answeredAt: '2026-09-01T09:00:00.000Z',
      }),
    ])

    expect(grouped['未分组'].map((item) => item.id)).toEqual(['fresh', 'old'])
  })

  it('interleaves cancelled items by event time instead of dumping them last', () => {
    const grouped = groupHistoryDesc([
      q({
        id: 'ans',
        content: 'answered',
        groupLabel: '同一组',
        createdAt: '2026-09-01T01:00:00.000Z',
        answeredAt: '2026-09-01T01:10:00.000Z',
      }),
      q({
        id: 'can',
        content: 'cancelled later',
        groupLabel: '同一组',
        status: 'cancelled',
        answered: false,
        createdAt: '2026-09-01T02:00:00.000Z',
      }),
    ])

    expect(grouped['同一组'].map((item) => item.id)).toEqual(['can', 'ans'])
  })

  it('falls back to 未分组 when groupLabel is empty', () => {
    const grouped = groupHistoryDesc([
      q({
        id: 'x',
        groupLabel: '',
        createdAt: '2026-09-01T01:00:00.000Z',
      }),
    ])
    expect(Object.keys(grouped)).toEqual(['未分组'])
  })

  it('returns an empty object when there is no history', () => {
    expect(
      groupHistoryDesc([q({ id: 'p', status: 'pending', answered: false })]),
    ).toEqual({})
  })
})
