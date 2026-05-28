import { apiRequest } from '@/shared/api/client'
import type { TenantQuotaView } from './quota-types'

export function getTenantQuota(tenantId: string) {
  return apiRequest<{ quota: TenantQuotaView }>(`/api/v1/tenants/${tenantId}/quota`, { tenantId })
}

export function recalculateTenantUsage(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/usage/recalculate`, { method: 'POST', tenantId })
}
