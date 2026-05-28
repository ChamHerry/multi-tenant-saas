import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createAPIKey, listAPIKeys, revokeAPIKey } from './api-key-api'
import type { CreateAPIKeyInput } from './api-key-types'

export const apiKeyKeys = {
  list: (tenantId?: string) => ['api-keys', 'list', tenantId] as const,
}

export function useAPIKeys(tenantId?: string) {
  return useQuery({
    queryKey: apiKeyKeys.list(tenantId),
    queryFn: () => listAPIKeys(tenantId!),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useCreateAPIKey(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateAPIKeyInput) => createAPIKey(tenantId!, input),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: apiKeyKeys.list(tenantId) }),
  })
}

export function useRevokeAPIKey(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (apiKeyId: string) => revokeAPIKey(tenantId!, apiKeyId),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: apiKeyKeys.list(tenantId) }),
  })
}
