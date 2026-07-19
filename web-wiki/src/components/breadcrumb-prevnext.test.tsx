import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, cleanup } from '@testing-library/react'
import { Breadcrumb } from '#/components/breadcrumb'
import { PrevNext } from '#/components/prev-next'

// Mock TanStack Router Link to render as a simple <a> tag
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, params, children, ...props }: { to: string; params?: Record<string, string>; children: React.ReactNode }) => {
    let href = to
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        if (key === '_splat') {
          href = href.replace(/\$/, value)
        } else {
          href = href.replace(`$${key}`, value)
        }
      }
    }
    return <a href={href} {...props}>{children}</a>
  },
}))

// ── Breadcrumb Tests ──

describe('Breadcrumb', () => {
  beforeEach(() => {
    cleanup()
  })

  it('renders 3 items with correct text', () => {
    const items = [
      { title: 'Home', path: 'home' },
      { title: 'Docs', path: 'docs' },
      { title: 'Page', path: 'docs/page' },
    ]
    const { container } = render(<Breadcrumb items={items} wikiId="123" />)

    expect(container.textContent).toContain('Home')
    expect(container.textContent).toContain('Docs')
    expect(container.textContent).toContain('Page')
  })

  it('renders separators between items', () => {
    const items = [
      { title: 'A', path: 'a' },
      { title: 'B', path: 'b' },
      { title: 'C', path: 'c' },
    ]
    const { container } = render(<Breadcrumb items={items} wikiId="123" />)

    const separators = container.querySelectorAll('span.text-sea-ink-soft\\/40')
    expect(separators.length).toBe(2)
  })

  it('last item is not a link', () => {
    const items = [
      { title: 'A', path: 'a' },
      { title: 'B', path: 'b' },
    ]
    const { container } = render(<Breadcrumb items={items} wikiId="123" />)

    const spans = container.querySelectorAll('span[aria-current="page"]')
    expect(spans.length).toBe(1)
    expect(spans[0].textContent).toBe('B')
  })

  it('non-last items are links', () => {
    const items = [
      { title: 'A', path: 'a' },
      { title: 'B', path: 'b' },
    ]
    const { container } = render(<Breadcrumb items={items} wikiId="123" />)

    const links = container.querySelectorAll('a')
    expect(links.length).toBe(1)
    expect(links[0].textContent).toBe('A')
  })

  it('returns null when items is empty array', () => {
    const { container } = render(<Breadcrumb items={[]} wikiId="123" />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null when items is undefined', () => {
    const { container } = render(<Breadcrumb wikiId="123" />)
    expect(container.firstChild).toBeNull()
  })

  it('single item shows no separator and is plain text', () => {
    const items = [{ title: 'Only', path: 'only' }]
    const { container } = render(<Breadcrumb items={items} wikiId="123" />)

    expect(container.textContent).toContain('Only')
    const separators = container.querySelectorAll('span.text-sea-ink-soft\\/40')
    expect(separators.length).toBe(0)
  })
})

// ── PrevNext Tests ──

describe('PrevNext', () => {
  beforeEach(() => {
    cleanup()
  })

  it('renders both prev and next when both are provided', () => {
    const { container } = render(
      <PrevNext
        prev={{ title: 'Previous Page', path: 'prev' }}
        next={{ title: 'Next Page', path: 'next' }}
        wikiId="123"
      />,
    )

    expect(container.textContent).toContain('Previous Page')
    expect(container.textContent).toContain('Next Page')
    expect(container.textContent).toContain('上一页')
    expect(container.textContent).toContain('下一页')
  })

  it('renders only next when prev is null', () => {
    const { container } = render(
      <PrevNext
        prev={null}
        next={{ title: 'Next Only', path: 'next' }}
        wikiId="123"
      />,
    )

    expect(container.textContent).toContain('Next Only')
    expect(container.textContent).toContain('下一页')
    expect(container.textContent).not.toContain('上一页')
  })

  it('renders only prev when next is null', () => {
    const { container } = render(
      <PrevNext
        prev={{ title: 'Prev Only', path: 'prev' }}
        next={null}
        wikiId="123"
      />,
    )

    expect(container.textContent).toContain('Prev Only')
    expect(container.textContent).toContain('上一页')
    expect(container.textContent).not.toContain('下一页')
  })

  it('returns null when both prev and next are null', () => {
    const { container } = render(<PrevNext prev={null} next={null} wikiId="123" />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null when both prev and next are undefined', () => {
    const { container } = render(<PrevNext wikiId="123" />)
    expect(container.firstChild).toBeNull()
  })

  it('prev link navigates to correct path', () => {
    const { container } = render(
      <PrevNext
        prev={{ title: 'Prev', path: 'getting-started' }}
        next={null}
        wikiId="123"
      />,
    )

    const link = container.querySelector('a')
    expect(link).not.toBeNull()
    expect(link?.getAttribute('href')).toBe('/wiki/123/getting-started')
  })

  it('next link navigates to correct path', () => {
    const { container } = render(
      <PrevNext
        prev={null}
        next={{ title: 'Next', path: 'advanced/config' }}
        wikiId="123"
      />,
    )

    const link = container.querySelector('a')
    expect(link).not.toBeNull()
    expect(link?.getAttribute('href')).toBe('/wiki/123/advanced/config')
  })
})
