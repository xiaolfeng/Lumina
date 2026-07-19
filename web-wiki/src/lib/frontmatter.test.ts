import { describe, expect, it } from 'vitest'
import { extractToc, parseFrontmatter } from './frontmatter'

describe('parseFrontmatter', () => {
  it('parses frontmatter and returns body', () => {
    const raw = '---\ntitle: Overview\ndescription: Project overview page\nicon: FileText\n---\n\n# Project Overview\n\nThis is the body.'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toEqual({
      title: 'Overview',
      description: 'Project overview page',
      icon: 'FileText',
    })
    expect(result.body).toBe('\n# Project Overview\n\nThis is the body.')
  })

  it('returns null frontmatter when string does not start with ---\\n', () => {
    const raw = '# Just a heading\n\nNo frontmatter here.'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toBeNull()
    expect(result.body).toBe(raw)
  })

  it('returns null frontmatter when closing delimiter is missing', () => {
    const raw = '---\ntitle: Broken\n\n# Body starts early'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toBeNull()
    expect(result.body).toBe(raw)
  })

  it('handles empty frontmatter block', () => {
    const raw = '---\n---\n\nBody after empty frontmatter.'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toEqual({})
    expect(result.body).toBe('\nBody after empty frontmatter.')
  })

  it('ignores comment lines inside frontmatter', () => {
    const raw = '---\ntitle: Test\n# this is a comment\ndescription: Desc\n---\n\nBody.'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toEqual({
      title: 'Test',
      description: 'Desc',
    })
  })

  it('handles frontmatter with no trailing newline after closing ---', () => {
    const raw = '---\ntitle: X\n---\nBody without extra newline.'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toEqual({ title: 'X' })
    expect(result.body).toBe('Body without extra newline.')
  })
})

describe('extractToc', () => {
  it('extracts h2, h3, and h4 headings', () => {
    const markdown = `
# H1 — page title (ignored)

## Getting Started

Some text.

### Installation

More text.

#### Deep Dive

Even more.

## API Reference

Final section.
    `
    const items = extractToc(markdown)
    expect(items).toHaveLength(4)
    expect(items[0]).toEqual({ depth: 2, title: 'Getting Started', slug: 'getting-started' })
    expect(items[1]).toEqual({ depth: 3, title: 'Installation', slug: 'installation' })
    expect(items[2]).toEqual({ depth: 4, title: 'Deep Dive', slug: 'deep-dive' })
    expect(items[3]).toEqual({ depth: 2, title: 'API Reference', slug: 'api-reference' })
  })

  it('ignores headings inside code blocks', () => {
    const markdown = `
## Real Heading

\`\`\`markdown
## Fake Heading
\`\`\`

### Another Real

~~~text
### Also Fake
~~~
    `
    const items = extractToc(markdown)
    expect(items).toHaveLength(2)
    expect(items[0]).toEqual({ depth: 2, title: 'Real Heading', slug: 'real-heading' })
    expect(items[1]).toEqual({ depth: 3, title: 'Another Real', slug: 'another-real' })
  })

  it('strips inline markdown from heading text', () => {
    const markdown = '## `code` and **bold** and *italic*'
    const items = extractToc(markdown)
    expect(items).toHaveLength(1)
    expect(items[0].title).toBe('`code` and **bold** and *italic*')
    expect(items[0].slug).toBe('code-and-bold-and-italic')
  })

  it('returns empty array for content with no headings', () => {
    const markdown = 'Just some plain text.\n\nNo headings at all.'
    expect(extractToc(markdown)).toEqual([])
  })

  it('returns empty array for empty string', () => {
    expect(extractToc('')).toEqual([])
  })

  it('slugifies unicode and punctuation correctly', () => {
    const markdown = '## 快速开始 — Getting Started!\n### C++ 编程'
    const items = extractToc(markdown)
    expect(items).toHaveLength(2)
    expect(items[0].slug).toBe('快速开始-getting-started')
    expect(items[1].slug).toBe('c-编程')
  })
})
