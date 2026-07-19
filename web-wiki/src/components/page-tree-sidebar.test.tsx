import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SidebarProvider } from '@lumina/components/ui/sidebar'
import { PageTreeSidebar } from './page-tree-sidebar'
import type { ManifestResponse } from '#/lib/api-client'

// Mock motion/react to avoid animation issues in tests
vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: { children: React.ReactNode }) => (
      <div {...props}>{children}</div>
    ),
  },
}))

// Mock @tanstack/react-router Link
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: { children: React.ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}))

vi.mock('@lumina/components/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

// Mock @tanstack/react-query
const mockUseQuery = vi.fn()
vi.mock('@tanstack/react-query', () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}))

function makeManifest(): ManifestResponse {
  return {
    navigation: [
      {
        title: 'Overview',
        path: 'overview',
        icon: 'BookOpen',
      },
      {
        title: 'Modules',
        path: 'modules',
        icon: 'Folder',
        default_open: true,
        children: [
          {
            title: 'Auth',
            path: 'modules/auth',
            icon: 'Lock',
          },
          {
            title: 'API',
            path: 'modules/api',
            icon: 'Code',
          },
          {
            title: '',
            path: '',
            separator: '---Advanced---',
          },
          {
            title: 'Database',
            path: 'modules/database',
            icon: 'Database',
          },
        ],
      },
      {
        title: 'Guides',
        path: 'guides',
        icon: 'Compass',
        children: [
          {
            title: 'Getting Started',
            path: 'guides/getting-started',
            icon: 'Sparkles',
          },
        ],
      },
    ],
    home: 'overview',
    language: 'zh',
    project_name: 'TestWiki',
    meta: { title: 'TestWiki', description: 'A test wiki', icon: 'BookOpen' },
  }
}

function renderWithProvider(ui: React.ReactElement) {
  return render(<SidebarProvider>{ui}</SidebarProvider>)
}

describe('PageTreeSidebar', () => {
  beforeEach(() => {
    mockUseQuery.mockReturnValue({
      data: makeManifest(),
      isLoading: false,
      error: null,
    })
    localStorage.clear()
  })

  it('renders the sidebar with data-testid', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    const sidebars = screen.getAllByTestId('wiki-sidebar')
    expect(sidebars.length).toBeGreaterThan(0)
  })

  it('renders top-level menu items', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.queryAllByText('Overview').length).toBeGreaterThan(0)
    expect(screen.queryAllByText('Modules').length).toBeGreaterThan(0)
    expect(screen.queryAllByText('Guides').length).toBeGreaterThan(0)
  })

  it('renders nested menu items under expanded directories', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    // Modules has default_open: true, so children should be visible
    expect(screen.queryAllByText('Auth').length).toBeGreaterThan(0)
    expect(screen.queryAllByText('API').length).toBeGreaterThan(0)
    expect(screen.queryAllByText('Database').length).toBeGreaterThan(0)
  })

  it('renders separator nodes', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    const separators = document.querySelectorAll(
      '[data-sidebar="separator"]',
    )
    expect(separators.length).toBeGreaterThan(0)
  })

  it('highlights the current page path', () => {
    renderWithProvider(
      <PageTreeSidebar
        wikiId="test-wiki"
        currentPagePath="modules/auth"
      />,
    )
    const authItems = screen.queryAllByText('Auth')
    expect(authItems.length).toBeGreaterThan(0)
  })

  it('renders the project name from meta title', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.queryAllByText('TestWiki').length).toBeGreaterThan(0)
  })

  it('renders the powered-by footer', () => {
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.queryAllByText(/Lumina/).length).toBeGreaterThan(0)
  })

  it('shows loading state', () => {
    mockUseQuery.mockReturnValueOnce({
      data: null,
      isLoading: true,
      error: null,
    })
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.getByText('加载中...')).toBeDefined()
  })

  it('shows error state', () => {
    mockUseQuery.mockReturnValueOnce({
      data: null,
      isLoading: false,
      error: new Error('Network error'),
    })
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.getByText('Network error')).toBeDefined()
  })

  it('renders empty state when no navigation items', () => {
    mockUseQuery.mockReturnValueOnce({
      data: {
        navigation: [],
        home: '',
        language: 'zh',
        project_name: 'EmptyWiki',
      },
      isLoading: false,
      error: null,
    })
    renderWithProvider(<PageTreeSidebar wikiId="test-wiki" />)
    expect(screen.getByText('暂无页面')).toBeDefined()
  })
})
