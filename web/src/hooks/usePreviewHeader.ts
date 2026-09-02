import { createContext, useContext } from 'react'

export const PreviewHeaderContext = createContext<{
  title: string | null
  setTitle: (title: string | null) => void
}>({
  title: null,
  setTitle: () => {},
})

export function usePreviewHeader() {
  return useContext(PreviewHeaderContext)
}
