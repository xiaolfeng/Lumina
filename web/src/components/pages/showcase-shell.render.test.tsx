/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ShowcaseShell } from './showcase-shell'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}))

vi.mock('#/hooks/usePages', () => ({
  useForkPage: () => ({ mutate: vi.fn(), isPending: false }),
  useSwitchActiveVersion: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('#/components/preview/file-viewer', () => ({
  PreviewFileViewer: ({ filename }: { filename: string }) => (
    <div data-testid="file-viewer">{filename}</div>
  ),
}))

afterEach(() => {
  cleanup()
})

describe('ShowcaseShell UI', () => {
  const mockVersion = {
    id: 'ver_1',
    page_id: 'page_1',
    version: 'v1.0.0',
    changelog: '初始版本',
    entry_filename: 'index.html',
    file_count: 2,
    total_size: 1024,
    created_by: 'admin',
    is_active: true,
    created_at: '2026-09-16T12:00:00Z',
  }

  const mockPage = {
    id: 'page_1',
    project_id: 'proj_1',
    project_name: 'lumina',
    slug: 'design-system',
    title: '微明统一设计规范',
    description: '页面设计规范',
    status: 'published' as const,
    access_mode: 'public' as const,
    latest_version_id: 'ver_1',
    latest_version: 'v1.0.0',
    page_url: '/pages/lumina/design-system/',
    created_at: '2026-09-16T12:00:00Z',
    updated_at: '2026-09-16T12:00:00Z',
  }

  const mockFiles = [
    {
      id: 'f_1',
      version_id: 'ver_1',
      filename: 'index.html',
      mime_type: 'text/html',
      size: 512,
      created_at: '2026-09-16T12:00:00Z',
      updated_at: '2026-09-16T12:00:00Z',
    },
    {
      id: 'f_2',
      version_id: 'ver_1',
      filename: 'style.css',
      mime_type: 'text/css',
      size: 512,
      created_at: '2026-09-16T12:00:00Z',
      updated_at: '2026-09-16T12:00:00Z',
    },
  ]

  it('renders top Header and floating Bar with capsule header', () => {
    render(
      <ShowcaseShell
        projectName="lumina"
        slug="design-system"
        filepath="index.html"
        page={mockPage}
        version={mockVersion}
        files={mockFiles}
        versions={[mockVersion]}
      />,
    )

    // 顶部 Header
    expect(screen.getByText('Lumina')).toBeTruthy()
    expect(screen.getByText('Pages')).toBeTruthy()
    expect(
      screen.getByRole('heading', { name: '微明统一设计规范' }),
    ).toBeTruthy()

    // 悬浮 Bar
    const pillBar = screen.getByRole('button', {
      name: '展开或折叠页面管理卡片',
    })
    expect(pillBar).toBeTruthy()
    expect(screen.getByText('lumina / design-system')).toBeTruthy()
    expect(screen.getByText('v1.0.0')).toBeTruthy()

    // 初始状态下抽屉未展开，且没有右侧贴边 aside
    expect(screen.queryByRole('complementary')).toBeNull()
    expect(screen.queryByText('元信息')).toBeNull()

    // 单击悬浮 Bar，展开悬浮卡片面板
    fireEvent.click(pillBar)
    expect(screen.getByText('元信息')).toBeTruthy()
    expect(screen.getByText('可渲染页面')).toBeTruthy()
    expect(screen.getByText('复制当前路径')).toBeTruthy()

    // 再次点击悬浮 Bar 折叠面板
    fireEvent.click(pillBar)
    expect(screen.queryByText('元信息')).toBeNull()

    // 重新打开后，点击外部区域折叠面板
    fireEvent.click(pillBar)
    expect(screen.getByText('元信息')).toBeTruthy()
    fireEvent.pointerDown(document.body)
    expect(screen.queryByText('元信息')).toBeNull()
  })
})
