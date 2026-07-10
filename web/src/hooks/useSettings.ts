import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSettings, updateSettings } from '#/lib/apis/settings'
import type { BaseResponse } from '#/lib/models/response/common'
import type { CategorySettings } from '#/lib/apis/settings'

export function useSettings(category: string) {
  return useQuery<BaseResponse<CategorySettings>>({
    queryKey: ['settings', category],
    queryFn: () => getSettings(category),
    enabled: !!category,
  })
}

export function useUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      category,
      items,
    }: {
      category: string
      items: { key: string; value: string }[]
    }) => updateSettings(category, items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}
