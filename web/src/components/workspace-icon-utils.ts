import { iconNames } from 'lucide-react/dynamic'

export const WORKSPACE_SLUG_PATTERN = /^[a-z]([a-z0-9-]*[a-z0-9])?$/
export const WORKSPACE_SLUG_MAX_LENGTH = 63

const lucideIconNameSet = new Set<string>(iconNames)
const lucideIconByCompactName = new Map(
  iconNames.map((name) => [name.replaceAll('-', ''), name]),
)

const ASCII_ICON_TOKEN = /^[a-zA-Z0-9:_-]+$/
const KEYCAP_EMOJI = /^[0-9#*]\uFE0F?\u20E3$/u
const FLAG_EMOJI = /^\p{Regional_Indicator}{2}$/u
const EMOJI_CORE = /\p{Extended_Pictographic}|\p{Emoji_Presentation}/u
const EMOJI_RELATED_CHARS =
  /^(?:\p{Extended_Pictographic}|\p{Emoji_Presentation}|\p{Emoji_Modifier}|\p{Regional_Indicator}|\uFE0F|\uFE0E|\u200D|\u20E3|[\u{E0020}-\u{E007F}])+$/u
const COMPLETE_EMOJI_STRING =
  /^(?:(?:\p{Regional_Indicator}{2})|(?:[0-9#*]\uFE0F?\u20E3)|(?:\p{Extended_Pictographic}\uFE0F?\p{Emoji_Modifier}?(?:\u200D(?:\p{Regional_Indicator}{2}|[0-9#*]\uFE0F?\u20E3|\p{Extended_Pictographic}\uFE0F?\p{Emoji_Modifier}?))*))+$/u
const graphemeSegmenter =
  typeof Intl !== 'undefined' && typeof Intl.Segmenter === 'function'
    ? new Intl.Segmenter(undefined, { granularity: 'grapheme' })
    : null

export const FEATURED_WORKSPACE_ICONS = [
  { name: 'layout-grid', label: '网格' },
  { name: 'briefcase', label: '工作' },
  { name: 'home', label: '家' },
  { name: 'folder', label: '文件夹' },
  { name: 'building-2', label: '办公' },
  { name: 'map', label: '地图' },
  { name: 'key', label: '钥匙' },
  { name: 'book', label: '书本' },
  { name: 'code', label: '代码' },
  { name: 'sparkles', label: '灵感' },
  { name: 'heart', label: '心愿' },
  { name: 'star', label: '星标' },
  { name: 'users', label: '团队' },
  { name: 'globe', label: '全球' },
  { name: 'coffee', label: '咖啡' },
  { name: 'palette', label: '设计' },
  { name: 'cpu', label: '系统' },
  { name: 'rocket', label: '项目' },
  { name: 'leaf', label: '生活' },
  { name: 'camera', label: '影像' },
] as const

export type ParsedWorkspaceIcon =
  | { kind: 'lucide'; name: string }
  | { kind: 'emoji'; value: string }
  | { kind: 'fallback' }

export function toKebabIconName(value: string): string {
  return value
    .replaceAll('_', '-')
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/([A-Z])([A-Z][a-z])/g, '$1-$2')
    .replace(/([a-zA-Z])(\d)/g, '$1-$2')
    .replace(/(\d)([a-zA-Z])/g, '$1-$2')
    .toLowerCase()
}

export function isLucideIconName(name: string): boolean {
  return lucideIconNameSet.has(name)
}

function isEmojiGrapheme(grapheme: string): boolean {
  if (KEYCAP_EMOJI.test(grapheme) || FLAG_EMOJI.test(grapheme)) return true
  return EMOJI_CORE.test(grapheme) && EMOJI_RELATED_CHARS.test(grapheme)
}

export function isEmojiValue(value: string): boolean {
  if (!value || ASCII_ICON_TOKEN.test(value)) return false
  if (value.includes('://') || value.includes('/')) return false
  if (!graphemeSegmenter) return COMPLETE_EMOJI_STRING.test(value)
  const graphemes = [...graphemeSegmenter.segment(value)].map(
    (part) => part.segment,
  )
  return graphemes.length > 0 && graphemes.every(isEmojiGrapheme)
}

function stripLucidePrefix(value: string): string {
  return value.replace(/^lucide:/i, '').trim()
}

export function parseWorkspaceIcon(raw?: string | null): ParsedWorkspaceIcon {
  const trimmed = raw?.trim() ?? ''
  if (!trimmed || trimmed.includes('://') || trimmed.includes('/')) {
    return { kind: 'fallback' }
  }

  const withoutPrefix = stripLucidePrefix(trimmed)
  if (withoutPrefix) {
    const exactName = withoutPrefix.toLowerCase()
    if (lucideIconNameSet.has(exactName)) {
      return { kind: 'lucide', name: exactName }
    }
    const kebab = toKebabIconName(withoutPrefix)
    if (lucideIconNameSet.has(kebab)) {
      return { kind: 'lucide', name: kebab }
    }
    const compactName = lucideIconByCompactName.get(exactName)
    if (compactName) return { kind: 'lucide', name: compactName }
  }

  if (isEmojiValue(withoutPrefix) || isEmojiValue(trimmed)) {
    return { kind: 'emoji', value: withoutPrefix || trimmed }
  }

  return { kind: 'fallback' }
}

export function featuredWorkspaceIconLabel(name: string): string | undefined {
  return FEATURED_WORKSPACE_ICONS.find((item) => item.name === name)?.label
}

export function workspaceIconDisplayLabel(raw?: string | null): string {
  const trimmed = raw?.trim() ?? ''
  if (!trimmed) return '选择图标'
  const parsed = parseWorkspaceIcon(trimmed)
  if (parsed.kind === 'emoji') return '表情'
  if (parsed.kind === 'fallback') return '默认图标'
  return featuredWorkspaceIconLabel(parsed.name) ?? '已选图标'
}

export function workspaceIconPickerItemLabel(name: string): string {
  return featuredWorkspaceIconLabel(name) ?? name
}

export function searchLucideIconNames(query: string, limit = 80): string[] {
  const trimmed = query.trim()
  if (!trimmed) {
    return FEATURED_WORKSPACE_ICONS.filter((item) =>
      lucideIconNameSet.has(item.name),
    ).map((item) => item.name)
  }

  const withoutPrefix = stripLucidePrefix(trimmed)
  if (isEmojiValue(withoutPrefix) || isEmojiValue(trimmed)) {
    return []
  }

  const lowered = withoutPrefix.toLowerCase()
  const kebab = toKebabIconName(withoutPrefix)
  const results: string[] = []
  const seen = new Set<string>()

  const push = (name: string) => {
    if (!lucideIconNameSet.has(name) || seen.has(name)) return
    seen.add(name)
    results.push(name)
  }

  push(kebab)

  for (const item of FEATURED_WORKSPACE_ICONS) {
    if (
      item.name.includes(lowered) ||
      item.label.includes(trimmed) ||
      item.label.toLowerCase().includes(lowered)
    ) {
      push(item.name)
    }
  }

  for (const name of iconNames) {
    if (results.length >= limit) break
    if (name.includes(lowered)) push(name)
  }

  return results.slice(0, limit)
}

export function isValidWorkspaceSlug(slug: string): boolean {
  return (
    slug.length > 0 &&
    slug.length <= WORKSPACE_SLUG_MAX_LENGTH &&
    WORKSPACE_SLUG_PATTERN.test(slug)
  )
}

function randomUuid(): string {
  if (typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  if (typeof crypto.getRandomValues !== 'function') {
    throw new Error('secure random uuid is unavailable')
  }
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  bytes[6] = (bytes[6] & 0x0f) | 0x40
  bytes[8] = (bytes[8] & 0x3f) | 0x80
  const hex = Array.from(bytes, (byte) =>
    byte.toString(16).padStart(2, '0'),
  ).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`
}

export function createHiddenWorkspaceSlug(): string {
  const slug = `space-${randomUuid()}`.toLowerCase()
  if (!isValidWorkspaceSlug(slug)) {
    throw new Error('generated workspace slug is invalid')
  }
  return slug
}
