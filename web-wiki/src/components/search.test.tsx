import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act, cleanup } from '@testing-library/react'
import { SearchDialog, WikiSearch } from './search'
import type { PageNode } from '#/lib/source'

// ── Mocks ──

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: { children: React.ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}))

vi.mock('#/lib/api-client', () => ({
  wikiReaderApi: {
    getPage: vi.fn(),
  },
}))

// Mock Orama
const mockInsert = vi.fn()
const mockSearch = vi.fn()
const mockCreate = vi.fn()
const mockCreateTokenizer = vi.fn()

vi.mock('@orama/orama', () => ({
  create: (...args: unknown[]) => mockCreate(...args),
  insert: (...args: unknown[]) => mockInsert(...args),
  search: (...args: unknown[]) => mockSearch(...args),
}))

vi.mock('@orama/tokenizers/mandarin', () => ({
  createTokenizer: () => mockCreateTokenizer(),
}))

// ── Helpers ──

function makeLeaves(): PageNode[] {
  return [
    { path: 'overview', title: '概述', description: '项目概述' },
    { path: 'auth/login', title: '登录认证', description: '用户登录与认证流程' },
    { path: 'auth/register', title: '注册', description: '用户注册流程' },
    { path: 'api/endpoints', title: 'API 端点', description: 'API 接口文档' },
  ]
}

// ── Tests ──

describe('SearchDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreate.mockResolvedValue({ id: 'test-db' })
    mockCreateTokenizer.mockResolvedValue({ tokenize: (text: string) => text.split('') })
    // Mock requestIdleCallback to execute immediately in tests
    window.requestIdleCallback = vi.fn((cb: () => void) => {
      cb()
      return 1
    }) as unknown as typeof window.requestIdleCallback
    window.cancelIdleCallback = vi.fn()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    cleanup()
  })

  it('renders search input when open', () => {
    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    expect(screen.getAllByPlaceholderText('搜索页面...')[0]).toBeDefined()
  })

  it('shows placeholder when no query', () => {
    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    expect(screen.getByText('输入关键词搜索页面...')).toBeDefined()
  })

  it('performs search with debounce', async () => {
    const mockHits = [
      {
        id: '1',
        document: { title: '登录认证', description: '用户登录与认证流程', path: 'auth/login' },
      },
    ]
    mockSearch.mockResolvedValue({ hits: mockHits, count: 1 })

    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    const input = screen.getAllByPlaceholderText('搜索页面...')[0]
    await act(async () => {
      fireEvent.change(input, { target: { value: '认证' } })
    })

    // Wait for debounce
    await waitFor(() => {
      expect(mockSearch).toHaveBeenCalled()
    }, { timeout: 500 })
  })

  it('searches Chinese keyword "认证" and returns correct result', async () => {
    const mockHits = [
      {
        id: '1',
        document: {
          title: '登录认证',
          description: '用户登录与认证流程',
          path: 'auth/login',
        },
      },
    ]
    mockSearch.mockResolvedValue({ hits: mockHits, count: 1 })

    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    const input = screen.getAllByPlaceholderText('搜索页面...')[0]
    await act(async () => {
      fireEvent.change(input, { target: { value: '认证' } })
    })

    await waitFor(() => {
      expect(mockSearch).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          term: '认证',
          limit: 10,
        }),
      )
    }, { timeout: 500 })

    // Verify the search result is displayed
    await waitFor(() => {
      expect(screen.getByText('登录认证')).toBeDefined()
    }, { timeout: 500 })
  })

  it('shows no results message when search returns empty', async () => {
    mockSearch.mockResolvedValue({ hits: [], count: 0 })

    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    const input = screen.getAllByPlaceholderText('搜索页面...')[0]
    await act(async () => {
      fireEvent.change(input, { target: { value: '不存在的关键词' } })
    })

    await waitFor(() => {
      expect(screen.getByText('未找到结果')).toBeDefined()
    }, { timeout: 500 })
  })

  it('handles keyboard navigation with arrow keys', async () => {
    const mockHits = [
      {
        id: '1',
        document: { title: '结果1', description: '描述1', path: 'path1' },
      },
      {
        id: '2',
        document: { title: '结果2', description: '描述2', path: 'path2' },
      },
    ]
    mockSearch.mockResolvedValue({ hits: mockHits, count: 2 })

    render(
      <SearchDialog
        wikiId="test-wiki"
        leaves={makeLeaves()}
        open={true}
        onOpenChange={() => {}}
      />,
    )

    const input = screen.getAllByPlaceholderText('搜索页面...')[0]
    await act(async () => {
      fireEvent.change(input, { target: { value: '结果' } })
    })

    await waitFor(() => {
      expect(screen.getByText('结果1')).toBeDefined()
    }, { timeout: 500 })

    // Test arrow down
    await act(async () => {
      fireEvent.keyDown(input, { key: 'ArrowDown' })
    })

    // Test arrow up
    await act(async () => {
      fireEvent.keyDown(input, { key: 'ArrowUp' })
    })

    // Test Enter
    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter' })
    })
  })
})

describe('WikiSearch', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreate.mockResolvedValue({ id: 'test-db' })
    mockCreateTokenizer.mockResolvedValue({ tokenize: (text: string) => text.split('') })
  })

  afterEach(() => {
    vi.restoreAllMocks()
    cleanup()
  })

  it('renders search trigger button', () => {
    render(
      <WikiSearch wikiId="test-wiki" leaves={makeLeaves()} />,
    )

    expect(screen.getAllByText('搜索')[0]).toBeDefined()
  })

  it('opens dialog on click', async () => {
    render(
      <WikiSearch wikiId="test-wiki" leaves={makeLeaves()} />,
    )

    const button = screen.getAllByText('搜索')[0]
    await act(async () => {
      fireEvent.click(button)
    })

    expect(screen.getAllByPlaceholderText('搜索页面...')[0]).toBeDefined()
  })

  it('opens dialog on Cmd+K', async () => {
    render(
      <WikiSearch wikiId="test-wiki" leaves={makeLeaves()} />,
    )

    await act(async () => {
      fireEvent.keyDown(document, { key: 'k', metaKey: true })
    })

    expect(screen.getAllByPlaceholderText('搜索页面...')[0]).toBeDefined()
  })
})
