import { Check } from 'lucide-react'

import { Kicker, PanelCard } from './primitives'
import type { OptionItem, Question } from './types'

/** 将 Markdown 内容提取为纯文本摘要，用于历史列表展示 */
function plainTextSummary(md: string, maxLen = 120): string {
  const plain = md
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1') // 链接/图片 → 文本
    .replace(/```[\s\S]*?```/g, ' ') // 代码块
    .replace(/`([^`]+)`/g, '$1') // 行内代码
    .replace(/[#*~>_+\-=`]/g, '') // 语法标记
    .replace(/\n+/g, ' ') // 换行 → 空格
    .trim()
  return plain.length > maxLen ? plain.slice(0, maxLen) + '…' : plain
}

/** 将选项 ID 解析为可读 label，查不到时回退显示原始值 */
function labelOf(options: OptionItem[] | undefined, id: string): string {
  return options?.find((o) => o.id === id)?.label ?? id
}

/** 将各题型提交的 answer 格式化为可读字符串。
 *  对于存选项 ID 的题型（select/multi-select/options/rank/rate），
 *  通过 options 把雪花 ID 映射回 label；不传 options 则回退显示原始值。 */
function formatAnswer(answer: unknown, options?: OptionItem[]): string {
  if (answer == null) return '—'
  if (typeof answer === 'string') return answer
  if (typeof answer === 'number' || typeof answer === 'boolean')
    return String(answer)
  if (Array.isArray(answer)) return answer.join(', ')
  if (typeof answer === 'object') {
    const obj = answer as Record<string, unknown>
    // 单选/多选: { selected: string | string[] }，可选 other: string | string[]
    if ('selected' in obj && !('text' in obj)) {
      const rawSel = Array.isArray(obj.selected) ? obj.selected : [obj.selected]
      const selLabels = rawSel.map((s) =>
        s === '__other__' ? null : labelOf(options, String(s)),
      )
      // other 自定义文本（单选为 string、多选为 string[]）
      const otherRaw = obj.other
      const otherLabels =
        otherRaw == null
          ? []
          : (Array.isArray(otherRaw) ? otherRaw : [otherRaw]).map((s) =>
              String(s),
            )
      const parts = [
        ...selLabels.filter((s): s is string => s != null),
        ...otherLabels,
      ]
      return parts.length > 0 ? parts.join('、') : '—'
    }
    // options 题: { selected: string, feedback?: string }
    if ('feedback' in obj && 'selected' in obj) {
      const sel =
        String(obj.selected) === '__other__'
          ? '—'
          : labelOf(options, String(obj.selected))
      return obj.feedback ? `${sel}（${String(obj.feedback)}）` : sel
    }
    // 文本: { text: string }
    if ('text' in obj) return String(obj.text)
    // 代码: { code: string, language?: string } — 历史列表仅显示语言标签
    if ('code' in obj) {
      const lang = 'language' in obj ? String(obj.language) : ''
      return lang ? `[${lang}]` : '代码'
    }
    // 图片: { images: [{ filename, ... }] }
    if ('images' in obj && Array.isArray(obj.images)) {
      const names = obj.images
        .map((i) => (i && typeof i === 'object' && 'filename' in i ? String((i as Record<string, unknown>).filename) : ''))
        .filter(Boolean)
      return names.length > 0 ? `📷 ${names.join('、')}` : `📷 ${obj.images.length} 张图片`
    }
    // 文件: { files: [{ filename, ... }] }
    if ('files' in obj && Array.isArray(obj.files)) {
      const names = obj.files
        .map((f) => (f && typeof f === 'object' && 'filename' in f ? String((f as Record<string, unknown>).filename) : ''))
        .filter(Boolean)
      return names.length > 0 ? `📎 ${names.join('、')}` : `📎 ${obj.files.length} 个文件`
    }
    // 布尔: { choice: "yes" | "no" }
    if ('choice' in obj) return String(obj.choice)
    // 滑块: { value: number }
    if ('value' in obj) return String(obj.value)
    // 排序: { ranking: string[] }
    if ('ranking' in obj && Array.isArray(obj.ranking)) {
      return obj.ranking.map((id) => labelOf(options, String(id))).join(' → ')
    }
    // 评分: { ratings: Record<string, number> }
    if ('ratings' in obj && typeof obj.ratings === 'object') {
      return Object.entries(obj.ratings as Record<string, unknown>)
        .map(([k, v]) => `${labelOf(options, k)}: ${v}`)
        .join('、')
    }
    // 决策题 (diff/plan/review): { decision, feedback?, edited?, annotations? }
    if ('decision' in obj) {
      const decision = String(obj.decision)
      const labels: Record<string, string> = { approve: '批准', reject: '拒绝', edit: '已编辑', revise: '需修订' }
      const parts: string[] = [labels[decision] ?? decision]
      if ('feedback' in obj && obj.feedback) parts.push(String(obj.feedback))
      return parts.join('（') + (parts.length > 1 ? '）' : '')
    }
    // 兜底：过滤掉 content 等大字段后显示 JSON key 名
    const { content: _c, ...rest } = obj as Record<string, unknown>
    const keys = Object.keys(rest)
    if (keys.length === 0) return '—'
    return keys.map((k) => `${k}: ${typeof rest[k] === 'string' ? String(rest[k]) : JSON.stringify(rest[k])}`).join(', ')
  }
  return String(answer)
}

interface HistoryCardProps {
  answeredQuestions: Question[]
  groupedHistory: Record<string, Question[]>
}

export function HistoryCard({
  groupedHistory,
  answeredQuestions,
}: HistoryCardProps) {
  return (
    <PanelCard
      flushHeader
      header={
        <div className="px-4 py-2.5">
          <Kicker>历史问答</Kicker>
        </div>
      }
      bodyClassName="p-0"
    >
      <div className="space-y-0 divide-y divide-line/30">
        {Object.entries(groupedHistory).map(([group, questions]) => (
          <div key={group} className="px-3 py-2">
            <div className="mb-1.5 flex items-center gap-2">
              <span className="inline-flex items-center gap-1 rounded-full bg-lagoon/8 px-2 py-0.5 text-[10px] font-semibold text-lagoon-deep">
                {group}
              </span>
              <span className="text-[10px] text-sea-ink-soft">
                {questions.length} 个问答
              </span>
            </div>

            <div className="space-y-1.5">
              {questions.map((q) => (
                <div key={q.id} className="flex items-start gap-2">
                  <Check
                    className="mt-0.5 size-3.5 shrink-0 text-lagoon"
                    aria-hidden
                  />
                  <div className="min-w-0 flex-1">
                    <p
                      className="line-clamp-2 text-xs leading-relaxed text-sea-ink-soft"
                      title={q.content}
                    >
                      {plainTextSummary(q.content)}
                    </p>
                    <p className="mt-0.5 text-xs font-medium text-sea-ink">
                      → {formatAnswer(q.answer, q.options)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}

        {answeredQuestions.length === 0 && (
          <div className="px-4 py-6 text-center">
            <p className="text-xs text-sea-ink-soft/50">暂无历史记录</p>
          </div>
        )}
      </div>
    </PanelCard>
  )
}
