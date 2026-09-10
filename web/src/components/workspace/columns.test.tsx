/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { DataTable } from '#/components/data-table'
import type { WorkspaceItem } from '#/lib/models/response/workspace'
import { getWorkspaceColumns } from './columns'

afterEach(() => {
  cleanup()
})

const item: WorkspaceItem = {
  id: 'ws-1',
  name: '工作',
  slug: 'space-secret-identifier',
  description: '',
  icon: 'briefcase',
  is_default: false,
  created_at: '2026-01-02T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
}

describe('getWorkspaceColumns', () => {
  it('hides slug and icon identifiers in the list', () => {
    render(
      <DataTable
        columns={getWorkspaceColumns({
          onEdit: () => {},
          onDelete: () => {},
        })}
        data={[item]}
      />,
    )

    expect(screen.getByText('工作')).toBeTruthy()
    expect(screen.queryByText('标识')).toBeNull()
    expect(screen.queryByText('space-secret-identifier')).toBeNull()
    expect(screen.queryByText('briefcase')).toBeNull()
    expect(screen.getByRole('img', { name: '工作的图标' })).toBeTruthy()
  })
})
