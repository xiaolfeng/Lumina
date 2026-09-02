import { describe, expect, it } from 'vitest'

import {
  computeSessionProgress,
  sessionProgressPercent,
} from './session-progress'
import type { Question } from './types'

function q(
  partial: Partial<Question> & Pick<Question, 'id' | 'status'>,
): Question {
  return {
    sessionId: 's1',
    content: partial.id,
    type: 'text',
    groupLabel: '',
    answered: partial.status === 'answered',
    createdAt: '2026-09-01T00:00:00.000Z',
    ...partial,
  }
}

describe('computeSessionProgress', () => {
  it('counts answered, pending, and total including skipped/cancelled', () => {
    expect(
      computeSessionProgress([
        q({ id: 'a', status: 'answered' }),
        q({ id: 'b', status: 'answered' }),
        q({ id: 'c', status: 'pending', answered: false }),
        q({ id: 'd', status: 'skipped', answered: false }),
        q({ id: 'e', status: 'cancelled', answered: false }),
      ]),
    ).toEqual({ total: 5, answered: 2, remaining: 1 })
  })

  it('returns zeros for an empty list', () => {
    expect(computeSessionProgress([])).toEqual({
      total: 0,
      answered: 0,
      remaining: 0,
    })
  })
})

describe('sessionProgressPercent', () => {
  it('rounds answered over total', () => {
    expect(
      sessionProgressPercent({ total: 3, answered: 1, remaining: 2 }),
    ).toBe(33)
    expect(
      sessionProgressPercent({ total: 4, answered: 2, remaining: 2 }),
    ).toBe(50)
  })

  it('is 0 when there are no questions', () => {
    expect(
      sessionProgressPercent({ total: 0, answered: 0, remaining: 0 }),
    ).toBe(0)
  })
})
