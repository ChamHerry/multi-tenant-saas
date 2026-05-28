import { apiRequest } from '@/shared/api/client'
import type { AuditExportResponse, AuditLogListResponse } from './audit-types'

function qs(params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') query.set(key, String(value))
  })
  const value = query.toString()
  return value ? `?${value}` : ''
}

export function listTenantAuditLogs(tenantId: string, params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AuditLogListResponse>(`/api/v1/tenants/${tenantId}/audit-logs${qs(params)}`, { tenantId })
}

export function exportTenantAuditLogs(tenantId: string, params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AuditExportResponse>(`/api/v1/tenants/${tenantId}/audit-logs/export`, { method: 'POST', tenantId, body: params })
}

export function listSecurityEvents(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AuditLogListResponse>(`/api/v1/me/security-events${qs(params)}`, { skipTenant: true })
}

export function listPlatformAuditLogs(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AuditLogListResponse>(`/api/v1/admin/audit-logs${qs(params)}`, { skipTenant: true })
}
