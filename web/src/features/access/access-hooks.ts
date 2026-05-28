import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { getAccessSnapshot } from './access-api'
import { hasAll, hasAny } from './can'
import type { PlatformPermission, TenantAccess, TenantPermission } from './access-types'

export const accessKeys = {
  snapshot: ['access', 'snapshot'] as const,
}

export function useAccess() {
  return useQuery({
    queryKey: accessKeys.snapshot,
    queryFn: getAccessSnapshot,
    retry: false,
  })
}

export function useTenantAccess(tenantId?: string) {
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const access = useAccess()
  const resolvedTenantId = tenantId ?? currentTenantId
  const tenantAccess = useMemo(
    () => access.data?.tenants.find((tenant) => tenant.tenant_id === resolvedTenantId),
    [access.data?.tenants, resolvedTenantId],
  )
  return { ...access, tenantAccess, tenantId: resolvedTenantId }
}

export function useCanTenant(permissions: TenantPermission | TenantPermission[], tenantId?: string, mode: 'all' | 'any' = 'all') {
  const { tenantAccess } = useTenantAccess(tenantId)
  const required = Array.isArray(permissions) ? permissions : [permissions]
  return mode === 'all' ? hasAll(tenantAccess?.permissions, required) : hasAny(tenantAccess?.permissions, required)
}

export function useCanPlatform(permissions: PlatformPermission | PlatformPermission[], mode: 'all' | 'any' = 'all') {
  const access = useAccess()
  const required = Array.isArray(permissions) ? permissions : [permissions]
  const owned = access.data?.platform_admin?.permissions
  return mode === 'all' ? hasAll(owned, required) : hasAny(owned, required)
}

export function tenantCan(tenantAccess: TenantAccess | undefined, permissions: readonly TenantPermission[] = [], mode: 'all' | 'any' = 'all') {
  return mode === 'all' ? hasAll(tenantAccess?.permissions, permissions) : hasAny(tenantAccess?.permissions, permissions)
}

export function platformCan(permissions: readonly PlatformPermission[] | undefined, required: readonly PlatformPermission[] = [], mode: 'all' | 'any' = 'all') {
  return mode === 'all' ? hasAll(permissions, required) : hasAny(permissions, required)
}
