import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getTenantQuota, recalculateTenantUsage } from './quota-api'

export const quotaKeys = {
  detail: (tenantId?: string) => ['quota', 'detail', tenantId] as const,
}

export function useTenantQuota(tenantId?: string) {
  return useQuery({ queryKey: quotaKeys.detail(tenantId), queryFn: () => getTenantQuota(tenantId!), enabled: Boolean(tenantId), retry: false })
}

export function useRecalculateTenantUsage(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: () => recalculateTenantUsage(tenantId!), onSuccess: () => void queryClient.invalidateQueries({ queryKey: quotaKeys.detail(tenantId) }) })
}
