import { useQuery } from '@tanstack/react-query'
import * as api from '#/lib/apis/dashboard'

export function useDashboardOverview() {
  return useQuery({
    queryKey: ['dashboard', 'overview'],
    queryFn: () => api.getDashboardOverview(),
  })
}
