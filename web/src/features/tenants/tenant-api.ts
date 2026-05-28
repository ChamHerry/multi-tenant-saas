import { apiRequest } from '@/shared/api/client'
import type { CreateTenantInput, Tenant, TenantContext, UpdateTenantInput } from './tenant-types'

export function createTenant(input: CreateTenantInput) {
  return apiRequest<{ tenant: Tenant }>('/api/v1/tenants', {
    method: 'POST',
    body: input,
    skipTenant: true,
  })
}

export function getTenant(tenantId: string) {
  return apiRequest<{ tenant: Tenant }>(`/api/v1/tenants/${tenantId}`, { tenantId })
}

export function updateTenant(tenantId: string, input: UpdateTenantInput) {
  return apiRequest<{ tenant: Tenant }>(`/api/v1/tenants/${tenantId}`, {
    method: 'PATCH',
    tenantId,
    body: input,
  })
}

export function suspendTenant(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/suspend`, { method: 'POST', tenantId })
}

export function restoreTenant(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/restore`, { method: 'POST', tenantId })
}

export function deleteTenant(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}`, { method: 'DELETE', tenantId })
}

export function getTenantContext(tenantId?: string) {
  return apiRequest<{ tenant_context: TenantContext }>('/api/v1/tenant-context', { tenantId })
}
