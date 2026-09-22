import type { Question } from './types'

export interface SessionProgress {
  total: number
  answered: number
  cancelled: number
  skipped: number
  remaining: number
}

export function computeSessionProgress(questions: Question[]): SessionProgress {
  // 四桶之和恒等于 total：pending→remaining，cancelled/skipped/answered 各归各桶
  let answered = 0
  let cancelled = 0
  let skipped = 0
  let remaining = 0
  for (const q of questions) {
    if (q.status === 'pending') remaining += 1
    else if (q.status === 'cancelled') cancelled += 1
    else if (q.status === 'skipped') skipped += 1
    else answered += 1
  }
  return { total: questions.length, answered, cancelled, skipped, remaining }
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

export function sessionProgressSkippedPercent(
  progress: SessionProgress,
): number {
  if (progress.total <= 0) return 0
  return Math.round((progress.skipped / progress.total) * 100)
}
