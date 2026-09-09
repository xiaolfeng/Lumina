/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'
import type { WorkspaceItem } from '#/lib/models/response/workspace'
import {
  CURRENT_WORKSPACE_STORAGE_KEY,
  resetCurrentWorkspaceStoreForTests,
  useCurrentWorkspace,
} from './useCurrentWorkspace'

const { workspaces } = vi.hoisted(() => {
  const items: WorkspaceItem[] = [
    {
      id: 'default-id',
      name: '默认空间',
      slug: 'default',
      description: '',
      icon: 'home',
      is_default: true,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    },
    {
      id: 'work-id',
      name: '工作',
      slug: 'work',
      description: '',
      icon: 'briefcase',
      is_default: false,
      created_at: '2026-01-02T00:00:00Z',
      updated_at: '2026-01-02T00:00:00Z',
    },
  ]
  return { workspaces: items }
})

vi.mock('#/hooks/useWorkspace', () => ({
  useWorkspaceList: () => ({
    data: { data: { items: workspaces, total: 2 } },
    isLoading: false,
  }),
}))

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  useRouterState: () => '/console/project',
}))

function installLocalStorage() {
  const memory = new Map<string, string>()
  const localStorage = {
    getItem: (key: string) => memory.get(key) ?? null,
    setItem: (key: string, value: string) => {
      memory.set(key, value)
    },
    removeItem: (key: string) => {
      memory.delete(key)
    },
    clear: () => {
      memory.clear()
    },
  }
  Object.defineProperty(window, 'localStorage', {
    configurable: true,
    value: localStorage,
  })
}

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

function CurrentProbe({ testId }: { testId: string }) {
  const { current } = useCurrentWorkspace()
  return <div data-testid={testId}>{current?.id ?? 'none'}</div>
}

function Switcher() {
  const { current, workspaces: items, setCurrentWorkspace } = useCurrentWorkspace()
  return (
    <>
      <div data-testid="switcher">{current?.id ?? 'none'}</div>
      <button
        type="button"
        onClick={() => {
          const next = items.find((item) => item.id === 'work-id')
          if (next) setCurrentWorkspace(next)
        }}
      >
        switch-work
      </button>
    </>
  )
}

beforeEach(() => {
  installLocalStorage()
  resetCurrentWorkspaceStoreForTests()
})

afterEach(() => {
  cleanup()
})

describe('useCurrentWorkspace', () => {
  it('keeps the stored workspace instead of writing the default on first render', async () => {
    window.localStorage.setItem(CURRENT_WORKSPACE_STORAGE_KEY, 'work-id')
    resetCurrentWorkspaceStoreForTests()

    render(<CurrentProbe testId="current" />, { wrapper })

    await waitFor(() => {
      expect(screen.getByTestId('current').textContent).toBe('work-id')
    })
    expect(window.localStorage.getItem(CURRENT_WORKSPACE_STORAGE_KEY)).toBe(
      'work-id',
    )
  })

  it('shares the selected workspace across hook instances', async () => {
    render(
      <>
        <Switcher />
        <CurrentProbe testId="other" />
      </>,
      { wrapper },
    )

    await waitFor(() => {
      expect(screen.getByTestId('switcher').textContent).toBe('default-id')
    })
    expect(screen.getByTestId('other').textContent).toBe('default-id')

    fireEvent.click(screen.getByRole('button', { name: 'switch-work' }))

    await waitFor(() => {
      expect(screen.getByTestId('switcher').textContent).toBe('work-id')
    })
    expect(screen.getByTestId('other').textContent).toBe('work-id')
    expect(window.localStorage.getItem(CURRENT_WORKSPACE_STORAGE_KEY)).toBe(
      'work-id',
    )
  })
})
