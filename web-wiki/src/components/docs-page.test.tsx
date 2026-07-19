import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DocsPage } from './docs-page'
import type { PageResponse } from '#/lib/api-client'

// Mock motion/react to avoid animation issues in tests
vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: { children: React.ReactNode }) => (
      <div {...props}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}))

// Mock @tanstack/react-router Link
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: { children: React.ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}))

vi.mock('#/components/search', () => ({
  WikiSearch: ({ wikiId, leaves }: { wikiId: string; leaves: unknown[] }) => (
    <div data-testid="wiki-search">{wikiId} ({leaves.length})</div>
  ),
}))

// Mock use-mobile hook
vi.mock('@lumina/components/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

// Mock @tanstack/react-query
const mockUseQuery = vi.fn()
vi.mock('@tanstack/react-query', () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}))

function makePageData(overrides?: Partial<PageResponse>): PageResponse {
  return {
    path: 'overview',
    content: '# Hello\n\nThis is a test page.',
    title: 'Overview',
    description: 'Test description',
    ...overrides,
  }
}

describe('DocsPage', () => {
  beforeEach(() => {
    mockUseQuery.mockReturnValue({
      data: {
        navigation: [
          { title: 'Overview', path: 'overview' },
          { title: 'Guide', path: 'guide' },
        ],
        home: 'overview',
        language: 'zh',
        project_name: 'Test Wiki',
      },
      isLoading: false,
      error: null,
    })
  })

  it('sidebar DOM node is stable across page switches within same wiki', () => {
    // First render: path=A
    const pageDataA: PageResponse = makePageData({ path: 'overview' })
    const { rerender } = render(
      <DocsPage wikiId="wiki-1" pageData={pageDataA}>
        <div>Content A</div>
      </DocsPage>,
    )

    const nodeA = screen.getByTestId('wiki-sidebar')
    expect(nodeA).toBeDefined()

    // Rerender: path=B (same wiki)
    const pageDataB: PageResponse = makePageData({ path: 'guide' })
    rerender(
      <DocsPage wikiId="wiki-1" pageData={pageDataB}>
        <div>Content B</div>
      </DocsPage>,
    )

    const nodeB = screen.getByTestId('wiki-sidebar')
    expect(nodeB).toBeDefined()

    // Same DOM reference = sidebar did NOT remount
    expect(nodeA).toBe(nodeB)
  })

  it('renders page title and description from frontmatter', () => {
    const pageData: PageResponse = makePageData({
      title: 'Test Title',
      description: 'Test Description',
    })

    render(
      <DocsPage wikiId="wiki-1" pageData={pageData}>
        <div>Content</div>
      </DocsPage>,
    )

    expect(screen.getAllByText('Test Title').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Test Description')).toBeDefined()
  })

  it('renders breadcrumb when pageData.breadcrumb is provided', () => {
    const pageData: PageResponse = makePageData({
      breadcrumb: [
        { title: 'Home', path: 'home' },
        { title: 'Guide', path: 'guide' },
      ],
    })

    render(
      <DocsPage wikiId="wiki-1" pageData={pageData}>
        <div>Content</div>
      </DocsPage>,
    )

    expect(screen.getByText('Home')).toBeDefined()
    // 'Guide' may appear in both breadcrumb and sidebar, so use getAllByText
    expect(screen.getAllByText('Guide').length).toBeGreaterThanOrEqual(1)
  })

  it('renders last_updated when provided', () => {
    const pageData: PageResponse = makePageData({
      last_updated: 1704067200, // 2024-01-01 00:00:00 UTC
    })

    render(
      <DocsPage wikiId="wiki-1" pageData={pageData}>
        <div>Content</div>
      </DocsPage>,
    )

    expect(screen.getByText(/更新于/)).toBeDefined()
  })

  it('renders prev/next navigation when provided', () => {
    const pageData: PageResponse = makePageData({
      prev: { title: 'Previous Page', path: 'prev' },
      next: { title: 'Next Page', path: 'next' },
    })

    render(
      <DocsPage wikiId="wiki-1" pageData={pageData}>
        <div>Content</div>
      </DocsPage>,
    )

    expect(screen.getByText('Previous Page')).toBeDefined()
    expect(screen.getByText('Next Page')).toBeDefined()
  })
})
