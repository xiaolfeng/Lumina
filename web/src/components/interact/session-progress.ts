import type { Question } from './types'

export interface SessionProgress {
  total: number
  answered: number
  cancelled: number
  remaining: number
}

export function computeSessionProgress(questions: Question[]): SessionProgress {
  let answered = 0
  let cancelled = 0
  let remaining = 0
  for (const q of questions) {
    if (q.status === 'pending') remaining += 1
    else if (q.status === 'cancelled') cancelled += 1
    else if (q.status === 'answered' || q.answered) answered += 1
  }
  return { total: questions.length, answered, cancelled, remaining }
}

export function sessionProgressPercent(progress: SessionProgress): number {
  if (progress.total <= 0) return 0
  return Math.round((progress.answered / progress.total) * 100)
}

export function sessionProgressCancelledPercent(
  progress: SessionProgress,
): number {
  if (progress.total <= 0) return 0
  return Math.round((progress.cancelled / progress.total) * 100)
}
