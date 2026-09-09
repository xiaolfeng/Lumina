import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { useWorkspaceList } from '#/hooks/useWorkspace'
import type { WorkspaceItem } from '#/lib/models/response/workspace'

export const CURRENT_WORKSPACE_STORAGE_KEY = 'lumina.currentWorkspaceId'

function readStoredWorkspaceId(): string {
  if (typeof window === 'undefined') return ''
  try {
    return window.localStorage.getItem(CURRENT_WORKSPACE_STORAGE_KEY) ?? ''
  } catch {
    return ''
  }
}

function writeStoredWorkspaceId(id: string) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(CURRENT_WORKSPACE_STORAGE_KEY, id)
  } catch {
    // localStorage 不可用时只影响 UI 记忆，不影响写接口
  }
}

export function useCurrentWorkspace() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const { data, isLoading } = useWorkspaceList({ page: 1, size: 50 })
  const workspaces = data?.data?.items ?? []
  const [storedId, setStoredId] = useState('')

  useEffect(() => {
    setStoredId(readStoredWorkspaceId())
  }, [])

  const current = useMemo(() => {
    if (workspaces.length === 0) return null
    const matched = workspaces.find((item) => item.id === storedId)
    if (matched) return matched
    return workspaces.find((item) => item.is_default) ?? workspaces[0]
  }, [workspaces, storedId])

  useEffect(() => {
    if (!current) return
    if (storedId !== current.id) {
      writeStoredWorkspaceId(current.id)
      setStoredId(current.id)
    }
  }, [current, storedId])

  const setCurrentWorkspace = useCallback(
    (workspace: WorkspaceItem) => {
      writeStoredWorkspaceId(workspace.id)
      setStoredId(workspace.id)
      queryClient.invalidateQueries({ queryKey: ['project', 'list'] })
      queryClient.invalidateQueries({ queryKey: ['pin'] })
      queryClient.invalidateQueries({ queryKey: ['qa', 'sessions'] })
      queryClient.invalidateQueries({ queryKey: ['preview', 'sessions'] })

      if (
        pathname.startsWith('/console/project/') &&
        pathname !== '/console/project'
      ) {
        void navigate({ to: '/console/project' })
      }
    },
    [navigate, pathname, queryClient],
  )

  return {
    current,
    workspaces,
    isLoading,
    setCurrentWorkspace,
  }
}
