import { useCallback, useEffect, useMemo, useSyncExternalStore } from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { useWorkspaceOptions } from '#/hooks/useWorkspace'
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

let storedId = ''
const listeners = new Set<() => void>()

function emitStoredId() {
  for (const listener of listeners) listener()
}

function subscribeStoredId(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function getStoredId() {
  return storedId
}

function getServerStoredId() {
  return ''
}

function persistStoredId(id: string) {
  if (storedId === id) return
  storedId = id
  writeStoredWorkspaceId(id)
  emitStoredId()
}

function hydrateStoredId() {
  storedId = readStoredWorkspaceId()
}

if (typeof window !== 'undefined') {
  hydrateStoredId()
}

export function resetCurrentWorkspaceStoreForTests() {
  hydrateStoredId()
  emitStoredId()
}

export function useCurrentWorkspace() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const { data, isLoading } = useWorkspaceOptions()
  const workspaces = data ?? []
  const currentId = useSyncExternalStore(
    subscribeStoredId,
    getStoredId,
    getServerStoredId,
  )

  const current = useMemo(() => {
    if (workspaces.length === 0) return null
    const matched = workspaces.find((item) => item.id === currentId)
    if (matched) return matched
    return workspaces.find((item) => item.is_default) ?? workspaces[0]
  }, [workspaces, currentId])

  useEffect(() => {
    if (!current) return
    const known =
      currentId !== '' && workspaces.some((item) => item.id === currentId)
    if (known) return
    persistStoredId(current.id)
  }, [current, currentId, workspaces])

  const setCurrentWorkspace = useCallback(
    (workspace: WorkspaceItem) => {
      persistStoredId(workspace.id)
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
