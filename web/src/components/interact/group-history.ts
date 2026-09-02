import type { Question } from './types'

const UNGROUPED = '未分组'

/** 已回答 / 已取消进入历史区；pending / skipped 不展示。 */
export function isHistoryQuestion(q: Question): boolean {
  return q.status !== 'pending' && q.status !== 'skipped'
}

/**
 * 历史排序时间：优先 answeredAt（刚回答的乐观更新也写这个字段），
 * 取消/缺回答时间时回退 createdAt。
 */
export function historyEventTime(q: Question): number {
  const raw = q.answeredAt || q.createdAt
  const n = Date.parse(raw)
  return Number.isFinite(n) ? n : 0
}

/**
 * 按事件时间 DESC 分组。分组顺序取组内最新一条，组内条目同样新→旧。
 * 这样「最新问答」会出现在历史列表最上方，而不是卡在最老的分组下面。
 */
export function groupHistoryDesc(
  questions: Question[],
): Record<string, Question[]> {
  const sorted = questions.filter(isHistoryQuestion).sort((a, b) => {
    const diff = historyEventTime(b) - historyEventTime(a)
    if (diff !== 0) return diff
    return b.id.localeCompare(a.id)
  })

  const grouped: Record<string, Question[]> = {}
  for (const q of sorted) {
    const key = q.groupLabel || UNGROUPED
    if (!(key in grouped)) grouped[key] = []
    grouped[key].push(q)
  }
  return grouped
}
