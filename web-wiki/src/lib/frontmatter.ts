/**
 * Frontmatter / fenced-block utilities — simple string parsing (no gray-matter)
 *
 * Used as a client-side fallback for local preview or when the backend
 * returns raw .mdx content.
 */

export interface TocItem {
  depth: number
  title: string
  slug: string
}

/**
 * Parse YAML frontmatter from raw markdown content.
 *
 * Rules:
 * - Only recognises frontmatter when the string starts with `---\n` (exact bytes).
 * - Extracts everything between the opening `---\n` and the next `\n---\n`.
 * - If there is no closing delimiter, returns the entire content as body
 *   (frontmatter = null) — no error is thrown.
 * - Frontmatter is parsed as simple `key: value` lines (one level, no nested
 *   objects or arrays). Values are kept as strings.
 */
export function parseFrontmatter(raw: string): {
  frontmatter: Record<string, unknown> | null
  body: string
} {
  if (!raw.startsWith('---\n')) {
    return { frontmatter: null, body: raw }
  }

  // Look for the closing delimiter: a newline, then `---`, then another newline
  const closeIdx = raw.indexOf('\n---\n', 3)
  if (closeIdx === -1) {
    return { frontmatter: null, body: raw }
  }

  const fmBlock = raw.slice(4, closeIdx)
  const body = raw.slice(closeIdx + 5)

  const frontmatter: Record<string, unknown> = {}
  for (const line of fmBlock.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const colonIdx = trimmed.indexOf(':')
    if (colonIdx === -1) continue
    const key = trimmed.slice(0, colonIdx).trim()
    const value = trimmed.slice(colonIdx + 1).trim()
    if (key) {
      frontmatter[key] = value
    }
  }

  return { frontmatter, body }
}

/**
 * Extract h2 / h3 / h4 headings from markdown text.
 *
 * Ignores headings inside code blocks (``` / ~~~) by stripping those
 * blocks before matching.
 *
 * Depth is the number of `#` characters (2, 3 or 4).
 * Slug is a simple kebab-case transformation:
 *   lowercase → spaces to hyphens → strip punctuation.
 */
export function extractToc(markdown: string): TocItem[] {
  const items: TocItem[] = []

  // Strip fenced code blocks so headings inside them are ignored
  const stripped = markdown
    .replace(/```[\s\S]*?```/g, '')
    .replace(/~~~[\s\S]*?~~~/g, '')

  for (const line of stripped.split('\n')) {
    const m = /^(#{2,4})\s+(.+)$/.exec(line)
    if (!m) continue
    const depth = m[1].length
    const raw = m[2].trim()
    if (!raw) continue
    items.push({ depth, title: raw, slug: toKebabSlug(raw) })
  }

  return items
}

/**
 * Simple kebab-case slug:
 *   lowercase → spaces to hyphens → strip punctuation → collapse hyphens.
 */
function toKebabSlug(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}
