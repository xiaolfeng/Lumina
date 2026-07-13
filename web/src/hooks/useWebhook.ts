import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as webhookApi from '../lib/apis/webhook'

export function useWebhookConfig(configId: string) {
  return useQuery({
    queryKey: ['webhook-config', configId],
    queryFn: () => webhookApi.getWebhookConfig(configId),
  })
}

export function useUpdateWebhookBranches(configId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (branches: string[]) => webhookApi.updateWebhookBranches(configId, branches),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhook-config', configId] }),
  })
}

export function useRegenerateWebhook(configId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => webhookApi.regenerateWebhook(configId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhook-config', configId] }),
  })
}

export function useWebhookEvents(configId: string, page: number, size: number) {
  return useQuery({
    queryKey: ['webhook-events', configId, page, size],
    queryFn: () => webhookApi.listWebhookEvents(configId, page, size),
  })
}
