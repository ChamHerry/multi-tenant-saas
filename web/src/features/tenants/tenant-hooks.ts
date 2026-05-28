import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys } from '@/features/auth/auth-hooks'
import { useTenantStore } from './tenant-store'
import {
  createTenant,
  deleteTenant,
  getTenant,
  getTenantContext,
  restoreTenant,
  suspendTenant,
  updateTenant,
} from './tenant-api'
import type { CreateTenantInput, UpdateTenantInput } from './tenant-types'

export const tenantKeys = {
  detail: (tenantId?: string) => ['tenants', 'detail', tenantId] as const,
  context: (tenantId?: string) => ['tenants', 'context', tenantId] as const,
}

export function useTenantDetail(tenantId?: string) {
  return useQuery({
    queryKey: tenantKeys.detail(tenantId),
    queryFn: () => getTenant(tenantId!),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useTenantContext() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  return useQuery({
    queryKey: tenantKeys.context(tenantId),
    queryFn: () => getTenantContext(tenantId),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useCreateTenant() {
  const queryClient = useQueryClient()
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  return useMutation({
    mutationFn: (input: CreateTenantInput) => createTenant(input),
    onSuccess: (data) => {
      setCurrentTenantId(data.tenant.id)
      void queryClient.invalidateQueries({ queryKey: authKeys.tenants })
      void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
      void queryClient.invalidateQueries({ queryKey: tenantKeys.context(data.tenant.id) })
    },
  })
}

export function useUpdateTenant(tenantId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateTenantInput) => updateTenant(tenantId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tenantKeys.detail(tenantId) })
      void queryClient.invalidateQueries({ queryKey: authKeys.tenants })
      void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
    },
  })
}

export function useTenantAction(tenantId: string) {
  const queryClient = useQueryClient()
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: tenantKeys.detail(tenantId) })
    void queryClient.invalidateQueries({ queryKey: authKeys.tenants })
    void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
  }
  return {
    suspend: useMutation({ mutationFn: () => suspendTenant(tenantId), onSuccess: invalidate }),
    restore: useMutation({ mutationFn: () => restoreTenant(tenantId), onSuccess: invalidate }),
    remove: useMutation({ mutationFn: () => deleteTenant(tenantId), onSuccess: invalidate }),
  }
}
