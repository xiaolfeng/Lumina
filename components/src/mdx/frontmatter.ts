/**
 * frontmatter.ts
 *
 * 轻量且稳健的 YAML frontmatter 纯前端解析器。
 * 用于从 .mdx 文档中提取结构化元数据（title, description, icon, tags 等）并分离正文。
 */

export interface MdxFrontmatter {
  title?: string
  description?: string
  icon?: string
  tags?: string[]
  date?: string
  last_updated?: string
  [key: string]: unknown
}

export function parseFrontmatter(raw: string): {
  frontmatter: MdxFrontmatter | null
  body: string
} {
  if (!raw || typeof raw !== 'string') {
    return { frontmatter: null, body: '' }
  }

  // 必须以 ---\n 或 ---\r\n 开头
  if (!raw.startsWith('---\n') && !raw.startsWith('---\r\n')) {
    return { frontmatter: null, body: raw }
  }

  const delimiterLen = raw.startsWith('---\r\n') ? 5 : 4
  const searchSlice = raw.slice(delimiterLen)

  // 寻找闭合的分隔符（\n---\n 或 \n---\r\n 或末尾 \n---）
  const closeOffset = searchSlice.search(/\r?\n---(?:\r?\n|$)/)
  if (closeOffset === -1) {
    return { frontmatter: null, body: raw }
  }

  const fmBlock = searchSlice.slice(0, closeOffset)
  const afterMatch = searchSlice.slice(closeOffset).match(/^\r?\n---(?:\r?\n)?/)
  const matchLen = afterMatch ? afterMatch[0].length : 4
  const body = searchSlice.slice(closeOffset + matchLen)

  const frontmatter: MdxFrontmatter = {}
  const lines = fmBlock.split(/\r?\n/)
  let currentListKey: string | null = null

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      continue
    }

    // 处理列表项（例如 tags 下的 - item）
    if (trimmed.startsWith('- ') && currentListKey) {
      const itemVal = cleanQuotes(trimmed.slice(2).trim())
      const list = frontmatter[currentListKey]
      if (Array.isArray(list)) {
        list.push(itemVal)
      } else {
        frontmatter[currentListKey] = [itemVal]
      }
      continue
    }

    // 处理普通 key: value
    const colonIdx = trimmed.indexOf(':')
    if (colonIdx === -1) {
      currentListKey = null
      continue
    }

    const key = trimmed.slice(0, colonIdx).trim()
    const rawVal = trimmed.slice(colonIdx + 1).trim()

    if (!key) continue

    if (!rawVal) {
      // 可能是接下来的列表开头（如 tags:）
      currentListKey = key
      frontmatter[key] = []
      continue
    }

    currentListKey = null

    // 解析内联数组，例如 [tag1, tag2]
    if (rawVal.startsWith('[') && rawVal.endsWith(']')) {
      const inner = rawVal.slice(1, -1).trim()
      const items = inner
        ? inner.split(',').map((it) => cleanQuotes(it.trim())).filter(Boolean)
        : []
      frontmatter[key] = items
      continue
    }

    // 布尔值与数值处理
    if (rawVal === 'true') {
      frontmatter[key] = true
    } else if (rawVal === 'false') {
      frontmatter[key] = false
    } else {
      frontmatter[key] = cleanQuotes(rawVal)
    }
  }

  // 特殊规整：如果 tags 是单个字符串，规整为数组
  if (typeof frontmatter.tags === 'string') {
    frontmatter.tags = [frontmatter.tags]
  }

  return { frontmatter, body }
}

function cleanQuotes(val: string): string {
  if (
    (val.startsWith('"') && val.endsWith('"')) ||
    (val.startsWith("'") && val.endsWith("'"))
  ) {
    return val.slice(1, -1)
  }
  return val
}
