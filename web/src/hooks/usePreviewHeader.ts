import { createContext, useContext } from 'react'
import type { WorkbenchDevice } from '#/components/preview/workbench-canvas'

export interface PreviewHeaderControls {
  device: WorkbenchDevice
  setDevice: (device: WorkbenchDevice) => void
  sourceMode: boolean
  setSourceMode: (mode: boolean) => void
  onPromoteClick: () => void
  sourcePageSlug?: string | null
}

export interface PreviewHeaderContextValue {
  title: string | null
  setTitle: (title: string | null) => void
  controls: PreviewHeaderControls | null
  setControls: (controls: PreviewHeaderControls | null) => void
}

export const PreviewHeaderContext = createContext<PreviewHeaderContextValue>({
  title: null,
  setTitle: () => {},
  controls: null,
  setControls: () => {},
})

export function usePreviewHeader() {
  return useContext(PreviewHeaderContext)
}
