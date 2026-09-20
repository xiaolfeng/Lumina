/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { PagesPasswordGate } from './password-gate'

const mockMutate = vi.fn()
let mockIsPending = false
let mockIsError = false
let mockError: Error | null = null

vi.mock('#/hooks/usePages', () => ({
  useUnlockPage: () => ({
    mutate: mockMutate,
    isPending: mockIsPending,
    isError: mockIsError,
    error: mockError,
  }),
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  mockIsPending = false
  mockIsError = false
  mockError = null
})

describe('PagesPasswordGate', () => {
  it('renders password protection prompt and submit button', () => {
    render(<PagesPasswordGate projectName="lumina" slug="demo" />)
    expect(screen.getByText('该页面受密码保护')).toBeTruthy()
    expect(screen.getByText('输入访问密码后继续浏览')).toBeTruthy()
    expect(screen.getByPlaceholderText('访问密码')).toBeTruthy()
    expect(screen.getByRole('button', { name: '解锁' })).toBeTruthy()
  })

  it('submits entered password via unlock mutation', () => {
    render(<PagesPasswordGate projectName="lumina" slug="demo" />)
    const input = screen.getByPlaceholderText('访问密码')
    fireEvent.change(input, { target: { value: 'my-secret' } })
    const button = screen.getByRole('button', { name: '解锁' })
    fireEvent.click(button)
    expect(mockMutate).toHaveBeenCalledWith({
      projectName: 'lumina',
      slug: 'demo',
      password: 'my-secret',
    })
  })

  it('displays inline error message when unlock fails', () => {
    mockIsError = true
    mockError = new Error('页面密码错误')
    render(<PagesPasswordGate projectName="lumina" slug="demo" />)
    expect(screen.getByText('页面密码错误')).toBeTruthy()
  })

  it('disables input and displays loading text while pending', () => {
    mockIsPending = true
    render(<PagesPasswordGate projectName="lumina" slug="demo" />)
    const input = screen.getByPlaceholderText('访问密码')
    expect((input as HTMLInputElement).disabled).toBe(true)
    expect(screen.getByRole('button', { name: '解锁中…' })).toBeTruthy()
  })
})
