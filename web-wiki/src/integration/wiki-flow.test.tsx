/**
 * 端到端集成测试 — Wiki Reader 全流程
 *
 * 覆盖：
 * 1. PasswordGate → DocsPage → Sidebar/Article/TOC 全链路渲染
 * 2. manifest 含 frontmatter + separator + icon
 * 3. page API 含 frontmatter + prev/next/breadcrumb
 * 4. 切换页面时 sidebar DOM 节点稳定（motion key=wikiId 修复）
 * 5. Cmd+K 打开搜索对话框
 *
 * 策略：使用真实 QueryClientProvider，仅 mock axios 层（wikiReaderApi）
 * 与 Orama 动态导入，以最大化集成覆盖度。
 */
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, act, cleanup, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { PasswordGate } from '#/components/password-gate'
import { DocsPage } from '#/components/docs-page'
import type { ManifestResponse, PageResponse } from '#/lib/api-client'

// ── Mocks ──

// Mock motion/react — 避免 framer-motion 在 jsdom 中的动画问题
vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: { children: ReactNode }) => (
      <div {...props}>{children}</div>
    ),
  },
  AnimatePresence: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}))

// Mock @tanstack/react-router Link — 渲染为普通 <a>
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: { children: ReactNode }) => (
    <a {...props}>{children}</a>
  ),
}))

// Mock @lumina/components/hooks/use-mobile
vi.mock('@lumina/components/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

// Mock Orama（搜索索引懒加载）
vi.mock('@orama/orama', () => ({
  create: vi.fn().mockResolvedValue({ id: 'test-db' }),
  insert: vi.fn().mockResolvedValue(undefined),
  search: vi.fn().mockResolvedValue({ hits: [], count: 0 }),
}))

vi.mock('@orama/tokenizers/mandarin', () => ({
  createTokenizer: vi.fn().mockResolvedValue({
    tokenize: (text: string) => text.split(''),
  }),
}))

// Mock wikiReaderApi（axios 层）— 集成测试的唯一 API 边界
const mockGetManifest = vi.fn()
const mockGetPage = vi.fn()
const mockCheckAuth = vi.fn()
const mockAuth = vi.fn()

vi.mock('#/lib/api-client', () => ({
  wikiReaderApi: {
    getManifest: (...args: unknown[]) => mockGetManifest(...args),
    getPage: (...args: unknown[]) => mockGetPage(...args),
    checkAuth: (...args: unknown[]) => mockCheckAuth(...args),
    auth: (...args: unknown[]) => mockAuth(...args),
  },
}))

// ── Fixtures ──

function makeManifest(): ManifestResponse {
  return {
    navigation: [
      {
        title: 'Overview',
        path: 'overview',
        icon: 'BookOpen',
        description: '项目概述',
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
            description: '认证模块',
          },
          {
            title: 'API',
            path: 'modules/api',
            icon: 'Code',
            description: 'API 接口',
          },
          // separator 节点
          {
            title: '',
            path: '',
            separator: '---Advanced---',
          },
          {
            title: 'Database',
            path: 'modules/database',
            icon: 'Database',
            description: '数据库设计',
          },
        ],
      },
    ],
    home: 'overview',
    language: 'zh',
    project_name: 'LuminaWiki',
    meta: {
      title: 'LuminaWiki',
      description: 'Lumina 项目文档',
      icon: 'BookOpen',
    },
  }
}

function makePageData(overrides?: Partial<PageResponse>): PageResponse {
  return {
    path: 'overview',
    content:
      '# Overview\n\n项目简介。\n\n## 架构设计\n\n架构内容。\n\n## 模块说明\n\n模块内容。',
    title: 'Overview',
    description: 'Lumina 项目总体概述',
    icon: 'BookOpen',
    last_updated: 1704067200,
    ...overrides,
  }
}

// ── Test Harness ──

function renderWithProviders(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  })
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  return render(ui, { wrapper })
}

/**
 * 渲染并返回自定义 rerender，确保 rerender 也包裹 QueryClientProvider。
 * （@testing-library/react v16 的 rerender 在某些场景下不复用 wrapper。）
 */
function renderWithProvidersAndRerender(initial: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  })
  const tree = (inner: ReactNode) => (
    <QueryClientProvider client={queryClient}>{inner}</QueryClientProvider>
  )
  const result = render(tree(initial))
  const customRerender = (next: ReactNode) => result.rerender(tree(next))
  return { ...result, rerender: customRerender }
}

// ── Tests ──

describe('Wiki Reader end-to-end flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    // 默认：无需密码
    mockCheckAuth.mockResolvedValue({
      password_required: false,
      authenticated: true,
    })
    // 默认 manifest
    mockGetManifest.mockResolvedValue(makeManifest())
    // 默认 page 数据
    mockGetPage.mockResolvedValue(makePageData())
    // requestIdleCallback 立即执行（搜索索引构建）
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

  it('renders full flow: PasswordGate → DocsPage → Sidebar/Article/TOC', async () => {
    const pageData = makePageData()

    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={pageData}>
          <div>Article body content</div>
        </DocsPage>
      </PasswordGate>,
    )

    // Sidebar 存在
    const sidebar = await screen.findByTestId('wiki-sidebar')
    expect(sidebar).toBeDefined()

    // Sidebar 包含 manifest 中的导航项（含 icon 节点）
    const overviewItems = await screen.findAllByText('Overview')
    expect(overviewItems.length).toBeGreaterThan(0)
    expect(screen.getAllByText('Modules').length).toBeGreaterThan(0)
    // Modules default_open=true，子项可见
    expect(screen.getAllByText('Auth').length).toBeGreaterThan(0)
    expect(screen.getAllByText('API').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Database').length).toBeGreaterThan(0)

    // Sidebar 顶部项目名（来自 meta.title）
    expect(screen.getAllByText('LuminaWiki').length).toBeGreaterThan(0)

    // Article 标题 + 描述（来自 frontmatter）
    expect(screen.getAllByText('Overview').length).toBeGreaterThan(0)
    expect(screen.getByText('Lumina 项目总体概述')).toBeDefined()

    // Article body 渲染
    expect(screen.getByText('Article body content')).toBeDefined()

    // 更新时间
    expect(screen.getByText(/更新于/)).toBeDefined()

    // TOC 渲染（content 含 ## 标题）
    expect(screen.getByText('本页目录')).toBeDefined()
    expect(screen.getByText('架构设计')).toBeDefined()
    expect(screen.getByText('模块说明')).toBeDefined()
  })

  it('renders separator nodes in sidebar', async () => {
    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={makePageData()}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    // 等待 manifest 加载完成（子项出现表示已渲染）
    await screen.findAllByText('Database')
    const separators = document.querySelectorAll('[data-sidebar="separator"]')
    expect(separators.length).toBeGreaterThan(0)
  })

  it('renders breadcrumb and prev/next from pageData', async () => {
    const pageData = makePageData({
      breadcrumb: [
        { title: 'Home', path: 'overview' },
        { title: 'Modules', path: 'modules' },
        { title: 'Auth', path: 'modules/auth' },
      ],
      prev: { title: 'Previous Page', path: 'overview' },
      next: { title: 'Next Page', path: 'modules/auth' },
    })

    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={pageData}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    await screen.findByTestId('wiki-sidebar')

    // Breadcrumb
    expect(screen.getByText('Home')).toBeDefined()
    expect(screen.getAllByText('Modules').length).toBeGreaterThan(0)
    // 最后一项 aria-current="page"
    const lastCrumb = screen.getByText('Auth')
    expect(lastCrumb).toBeDefined()

    // Prev/Next
    expect(screen.getByText('Previous Page')).toBeDefined()
    expect(screen.getByText('Next Page')).toBeDefined()
    expect(screen.getByText('上一页')).toBeDefined()
    expect(screen.getByText('下一页')).toBeDefined()
  })

  it('keeps sidebar DOM node stable across page switches within same wiki (motion key fix)', async () => {
    const pageA = makePageData({ path: 'overview' })
    const { rerender } = renderWithProvidersAndRerender(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={pageA}>
          <div>Content A</div>
        </DocsPage>
      </PasswordGate>,
    )

    const nodeA = await screen.findByTestId('wiki-sidebar')
    expect(nodeA).toBeDefined()

    const pageB = makePageData({
      path: 'modules/auth',
      title: 'Auth',
      description: '认证模块文档',
    })

    rerender(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={pageB}>
          <div>Content B</div>
        </DocsPage>
      </PasswordGate>,
    )

    const nodeB = screen.getByTestId('wiki-sidebar')
    expect(nodeB).toBeDefined()

    // 关键断言：DOM 节点引用稳定（motion key=wikiId 修复）
    expect(nodeA).toBe(nodeB)

    // 内容已切换
    expect(screen.getByText('Content B')).toBeDefined()
    expect(screen.queryByText('Content A')).toBeNull()
  })

  it('opens search dialog on Cmd+K', async () => {
    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={makePageData()}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    await screen.findByTestId('wiki-sidebar')

    // 触发 Cmd+K
    await act(async () => {
      fireEvent.keyDown(document, { key: 'k', metaKey: true })
    })

    // 搜索对话框打开
    await waitFor(() => {
      const inputs = screen.getAllByPlaceholderText('搜索页面...')
      expect(inputs.length).toBeGreaterThan(0)
    })
  })

  it('opens search dialog on button click', async () => {
    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={makePageData()}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    await screen.findByTestId('wiki-sidebar')

    const searchButtons = screen.getAllByText('搜索')
    expect(searchButtons.length).toBeGreaterThan(0)

    await act(async () => {
      fireEvent.click(searchButtons[0])
    })

    await waitFor(() => {
      const inputs = screen.getAllByPlaceholderText('搜索页面...')
      expect(inputs.length).toBeGreaterThan(0)
    })
  })

  it('shows password input when password required and not authenticated', async () => {
    mockCheckAuth.mockResolvedValue({
      password_required: true,
      authenticated: false,
    })

    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={makePageData()}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    // 应显示密码输入（PasswordInput 渲染密码字段）
    await waitFor(() => {
      const passwordInputs = screen.getAllByPlaceholderText(/密码|password/i)
      expect(passwordInputs.length).toBeGreaterThan(0)
    })
  })

  it('renders loading skeleton while auth-check is pending', async () => {
    // 让 checkAuth 永不 resolve，保持 loading
    mockCheckAuth.mockReturnValue(new Promise(() => {}))

    renderWithProviders(
      <PasswordGate wikiId="wiki-1">
        <DocsPage wikiId="wiki-1" pageData={makePageData()}>
          <div>body</div>
        </DocsPage>
      </PasswordGate>,
    )

    // 骨架屏存在（animate-pulse 元素）
    await waitFor(() => {
      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
    })
  })
})
