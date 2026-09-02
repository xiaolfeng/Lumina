import { createContext, useContext } from 'react'

import type { SessionProgress } from '#/components/interact/session-progress'

export type { SessionProgress }

export const SidebarOpenContext = createContext<{
  open: boolean
  setOpen: (v: boolean) => void
  progress: SessionProgress | null
  setProgress: (v: SessionProgress | null) => void
}>({
  open: false,
  setOpen: () => {},
  progress: null,
  setProgress: () => {},
})

export function useSidebarOpen() {
  return useContext(SidebarOpenContext)
}
