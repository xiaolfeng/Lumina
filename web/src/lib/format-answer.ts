/**
 * 通用回答格式化工具 —— 供控制台 QA 详情与 Interact 历史列表共用。
 * 将各题型提交的 answer（雪花 ID / 文本 / 结构化对象）格式化为可读字符串。
 */

/** 选项最小结构（兼容 interact 的 OptionItem 与 qa-admin 的 options 数组） */
export interface AnswerOption {
  id: string
  label?: string
  description?: string
}

/** 将选项 ID 解析为可读 label，查不到时回退显示原始值 */
function labelOf(options: AnswerOption[] | undefined, id: string): string {
  return options?.find((o) => o.id === id)?.label ?? id
}

/** 将各题型提交的 answer 格式化为可读字符串。
 *  对于存选项 ID 的题型（select/multi-select/options/rank/rate），
 *  通过 options 把雪花 ID 映射回 label；不传 options 则回退显示原始值。 */
export function formatAnswer(answer: unknown, options?: AnswerOption[]): string {
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
        .map((i) =>
          i && typeof i === 'object' && 'filename' in i
            ? String((i as Record<string, unknown>).filename)
            : '',
        )
        .filter(Boolean)
      return names.length > 0 ? `📷 ${names.join('、')}` : `📷 ${obj.images.length} 张图片`
    }
    // 文件: { files: [{ filename, ... }] }
    if ('files' in obj && Array.isArray(obj.files)) {
      const names = obj.files
        .map((f) =>
          f && typeof f === 'object' && 'filename' in f
            ? String((f as Record<string, unknown>).filename)
            : '',
        )
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
    return keys
      .map((k) => `${k}: ${typeof rest[k] === 'string' ? String(rest[k]) : JSON.stringify(rest[k])}`)
      .join(', ')
  }
  return String(answer)
}
