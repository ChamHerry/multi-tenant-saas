import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createTenantAPIKey,
  listTenantAPIKeys,
  revokeTenantAPIKey,
} from './tenant-api-key-api'
import type { CreateTenantAPIKeyInput } from './tenant-api-key-types'

export const tenantApiKeyKeys = {
  list: (tenantId?: string) => ['api-keys', 'tenant', tenantId] as const,
}

export function useTenantAPIKeys(tenantId?: string) {
  return useQuery({
    queryKey: tenantApiKeyKeys.list(tenantId),
    queryFn: () => listTenantAPIKeys(tenantId!),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useCreateTenantAPIKey(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTenantAPIKeyInput) => createTenantAPIKey(tenantId!, input),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: tenantApiKeyKeys.list(tenantId) }),
  })
}

export function useRevokeTenantAPIKey(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (apiKeyId: string) => revokeTenantAPIKey(tenantId!, apiKeyId),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: tenantApiKeyKeys.list(tenantId) }),
  })
}
